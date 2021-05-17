package effects

import (
	"fmt"
	"ledsim/internal"
	"math"
	"math/cmplx"
	"sync"
	"time"

	"github.com/1lann/dissonance/ffmpeg"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mjibson/go-dsp/fft"
)

type PitchEffect struct {
	mutex   *sync.Mutex
	colours []colorful.Color
}

// How many values we store in colour_coords so
// we have a smoother fade effect
const Buffer = 40
const SamplesPerSecond = 20
const Range = 48e3
const UpperFrequency = 1e4

const Threshold = 8e5

// Given an FFT transform, generates a colour that reflects the
// most common pitches
func generateColour(sample []complex128, threshold float64) colorful.Color {
	// Used to calculate the mean of points around circle
	colour := colorful.LuvLCh(0.5, 0, 0)
	peaks := 0

	for index, value := range sample {
		frequency := (index + 1) * SamplesPerSecond

		// Remove everything above the Nyquist frequency
		// (currently set to 10kHz)
		if frequency > UpperFrequency {
			break
		}

		abs := cmplx.Abs(value)
		if abs <= threshold {
			continue
		}

		// TODO: account for the height of each peak instead
		// of treating them all equally, thereby reducing static
		peaks += 1
		_, decimal := math.Modf(math.Log2(float64(frequency)))
		current_colour := colorful.LuvLCh(0.5, 1, decimal*360.0)

		if peaks == 1 {
			colour = current_colour
		} else {
			colour = colour.BlendLuvLCh(current_colour, 1.0/float64(peaks))
		}
	}

	return colour
}

func (e *PitchEffect) Apply(sys *internal.System, t time.Time) {
	for _, led := range sys.LEDs {
		// Fade effect logic
		distance := e.distance(led.X, led.Y)
		intpart, decimal := math.Modf(distance * Buffer)
		distance_index := int(intpart)

		// Stop distance_index from causing an out-of-bounds
		// exception if our distance is exactly 1
		if distance_index > Buffer-1 {
			distance_index -= 1
		}

		e.mutex.Lock()
		next_value := e.colours[distance_index+1]
		curr_value := e.colours[distance_index]
		e.mutex.Unlock()

		v := curr_value.BlendLuvLCh(next_value, decimal)

		// Borrowed from Jason's code
		led.R, led.G, led.B = v.Clamped().RGB255()
	}
}

// Given the x- and y-coordinates of an LED, returns a value
// between 0 and 1 based on its distance from a certain point
func (e *PitchEffect) distance(x, y float64) float64 {
	// The largest value of y after scaling
	const Largest = 3.263826198537793

	value := y / Largest

	if value < 0 || value > 1 {
		fmt.Println("Out of bounds!", x, y)
	}

	return y / Largest
}

func appendColour(arr []colorful.Color, value colorful.Color) []colorful.Color {
	rest := arr[1:]
	rest = append(rest, value)
	return rest
}

func NewPitchEffect(deviceName string, debug bool) *PitchEffect {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	empty_colours := make([]colorful.Color, 0, Buffer+1)
	for i := 0; i < Buffer+1; i++ {
		empty_colours = append(empty_colours, colorful.LuvLCh(0.5, 0, 0))
	}

	effect := &PitchEffect{
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
			colour := generateColour(fft_results, Threshold)

			// update the loudness
			effect.mutex.Lock()
			effect.colours = appendColour(effect.colours, colour)
			effect.mutex.Unlock()
		}
	}()

	return effect
}
