package main

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"sync"
	"time"

	"github.com/1lann/dissonance/audio"
	"github.com/1lann/dissonance/ffmpeg"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mjibson/go-dsp/fft"
)

type BasicMicFFT struct {
	mutex    *sync.Mutex
	loudness [][]complex128 // [0 .. 1]
}

func generateColour(sample []complex128) float64 {
	max_abs := float64(0)
	max_index := 0

	for index, value := range sample {
		abs := cmplx.Abs(value)

		if abs > max_abs {
			max_abs = abs
			max_index = index
		}
	}

	return 360.0 * float64(max_index) / float64(len(sample))
}

func (e *BasicMicFFT) Apply(sys *System, t time.Time) {
	for _, led := range sys.LEDs {
		distance := e.distance(led.X, led.Y)
		intpart, diff_multiplier := math.Modf(distance * Values)
		distance_index := int(intpart)

		if distance_index >= Values {
			distance_index = Values - 1
		}

		e.mutex.Lock()
		next_value := generateColour(e.loudness[(distance_index+1)%Values])
		curr_value := generateColour(e.loudness[distance_index])
		e.mutex.Unlock()

		if next_value < curr_value {
			next_value += 360.0
		}

		difference := next_value - curr_value
		v := curr_value + difference*diff_multiplier

		led.R, led.G, led.B = colorful.LuvLCh(0.5, 1, math.Mod(v, 360.0)).Clamped().RGB255()
	}
}

func (e *BasicMicFFT) distance(x, y float64) float64 {
	return (y + 2) / 4
}

func readFullFloat(stream audio.Stream, buf []float64) error {
	n := 0

	for n < len(buf) {
		read, err := stream.Read(buf[n:])
		n += read
		if err != nil {
			return err
		}
	}

	return nil
}

func appendRightFFT(arr [][]complex128, value []complex128) [][]complex128 {
	rest := arr[1:]
	rest = append(rest, value)
	return rest
}

func NewBasicMicFFT(deviceName string, debug bool) *BasicMicFFT {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	empty_loudness := make([][]complex128, 0, Values)
	for i := 0; i < Values; i++ {
		empty_loudness = append(empty_loudness, make([]complex128, 0, 20))
	}

	effect := &BasicMicFFT{
		mutex:    new(sync.Mutex),
		loudness: empty_loudness,
	}

	go func() {
		samplesToRead := stream.SampleRate() / 20
		buf := make([]float64, samplesToRead)

		log.Println("ok")

		for {
			// wait/read until we have 1/20th of a second worth of audio
			err := readFullFloat(stream, buf)
			if err != nil {
				panic(err)
			}

			log.Println("starting fft")
			t := time.Now()
			// average the amplitude
			fft_results := fft.FFTReal(buf)
			log.Println(time.Since(t))

			// update the loudness
			effect.mutex.Lock()
			effect.loudness = appendRightFFT(effect.loudness, fft_results)
			effect.mutex.Unlock()
		}
	}()

	return effect
}
