package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"sync"
	"time"

	"github.com/1lann/dissonance/ffmpeg"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mjibson/go-dsp/fft"
)

type MixingPitches struct {
	mutex   *sync.Mutex
	colours [][]colorful.Color // [TotalColours][0 .. 2*pi], NaN indicating no colour
}

// The number of colours we want to display
const TotalColours = 4
const Intervals = 40

func generateColours(sample []complex128, threshold float64) []colorful.Color {
	values := make([]float64, 0, TotalColours)
	indices := make([]int, 0, TotalColours)

	for i := 0; i < TotalColours; i++ {
		values = append(values, math.NaN())
		indices = append(indices, 0)
	}

	// Iterate over the range so we get the n largest
	// frequencies over our threshold, where n is TotalColours
	for index, value := range sample {
		// Remove everything above the Nyquist frequency
		// (currently set to 10kHz)
		if index*SamplesPerSecond > UpperFrequency {
			break
		}

		abs := cmplx.Abs(value)
		if abs <= threshold {
			continue
		}

		// Find the index that we have to replace everything at
		bigger_than := -1
		for j := 0; j < TotalColours; j++ {
			if math.IsNaN(values[j]) || abs >= values[j] {
				bigger_than = j
				break
			}
		}

		if bigger_than == -1 {
			continue
		}

		for k := TotalColours - 1; k > bigger_than; k-- {
			values[k-1] = values[k]
			indices[k-1] = indices[k]
		}

		values[bigger_than] = abs
		indices[bigger_than] = index
	}

	coords := make([]colorful.Color, 0, TotalColours)

	for l := 0; l < TotalColours; l++ {
		if math.IsNaN(values[l]) {
			coords = append(coords, colorful.LuvLCh(0.5, 0, 0))
		} else {
			frequency := indices[l] * SamplesPerSecond
			_, decimal := math.Modf(math.Log2(float64(frequency)))
			coords = append(coords, colorful.LuvLCh(0.5, 1, decimal*360.0))
		}
	}

	return coords
}

func (e *MixingPitches) Apply(sys *System, t time.Time) {
	for _, led := range sys.LEDs {
		colour := colorful.LuvLCh(0.5, 0, 0)
		divisor := 0

		for i := 0; i < TotalColours; i++ {
			// fmt.Println("Original point:", led.X, led.Y)
			distance := e.distance(led.X, led.Y, i)
			// fmt.Println("Distance from point", i, ": ", distance)
			intpart, decimal := math.Modf(distance * Intervals)
			distance_index := int(intpart)

			if distance_index >= Intervals {
				distance_index = Intervals - 1
			}

			e.mutex.Lock()
			next_value := e.colours[distance_index+1][i]
			curr_value := e.colours[distance_index][i]
			e.mutex.Unlock()

			current_colour := curr_value.BlendLuvLCh(next_value, decimal)
			is_grey := current_colour.AlmostEqualRgb(colorful.LuvLCh(0.5, 0, 0))

			if is_grey {
				continue
			}

			divisor++

			if divisor == 1 {
				colour = current_colour
			} else {
				colour = colour.BlendLuvLCh(current_colour, 1.0/float64(divisor))
			}
		}

		led.R, led.G, led.B = colour.Clamped().RGB255()
	}
}

func (e *MixingPitches) distance(x, y float64, index int) float64 {
	// We want to start the circle from the top and work our way
	// anticlockwise
	angle := (2 * math.Pi * float64(index) / TotalColours) + (math.Pi / 2)
	circle_x := 2 * math.Cos(angle)
	circle_y := 2 * math.Sin(angle)

	x_diff := x - circle_x
	y_diff := y - circle_y

	return math.Sqrt(x_diff*x_diff+y_diff*y_diff) / 4
}

func appendColours(arr [][]colorful.Color, value []colorful.Color) [][]colorful.Color {
	rest := arr[1:]
	rest = append(rest, value)
	return rest
}

func NewMixingPitches(deviceName string, debug bool) *MixingPitches {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	empty_colours := make([][]colorful.Color, 0, Buffer+1)
	for i := 0; i < Buffer+1; i++ {
		current_colours := make([]colorful.Color, 0, TotalColours)

		for j := 0; j < TotalColours; j++ {
			current_colours = append(current_colours, colorful.LuvLCh(0.5, 0, 0))
		}

		empty_colours = append(empty_colours, current_colours)
	}

	effect := &MixingPitches{
		mutex:   new(sync.Mutex),
		colours: empty_colours,
	}

	go func() {
		samplesToRead := stream.SampleRate() / SamplesPerSecond
		buf := make([]int16, samplesToRead)

		for {
			// wait/read until we have 1/20th of a second worth of audio
			err := readFull(stream, buf)
			if err != nil {
				panic(err)
			}

			buf64 := make([]float64, samplesToRead)

			for i := 0; i < samplesToRead; i++ {
				buf64[i] = float64(buf[i])
			}

			// fetch FFT results for the snippet (48k / 20 = 2.4k samples)
			fft_results := fft.FFTReal(buf64)
			colours := generateColours(fft_results, Threshold)

			// update the loudness
			effect.mutex.Lock()
			effect.colours = appendColours(effect.colours, colours)
			effect.mutex.Unlock()
		}
	}()

	return effect
}
