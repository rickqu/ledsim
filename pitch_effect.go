package main

import (
	"fmt"
	"math"
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

const Threshold = 6e5

func (e *PitchEffect) Apply(sys *System, t time.Time) {
	for _, led := range sys.LEDs {
		// Fade effect logic
		distance := e.distance(led.X, led.Y)
		intpart, decimal := math.Modf(distance * Buffer)
		distance_index := int(intpart)

		// Just to ensure we don't have any out-of-bounds exceptions
		if distance_index >= Buffer {
			distance_index = Buffer - 1
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

func (e *PitchEffect) distance(x, y float64) float64 {
	return (y + 2) / 4
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
