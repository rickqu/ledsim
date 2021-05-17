package effects

import (
	"fmt"
	"ledsim/internal"
	"math"
	"sync"
	"time"

	"github.com/1lann/dissonance/ffmpeg"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mjibson/go-dsp/fft"
)

const TimesPerSecond = 20
const Timeout = 1

const Intervals = 40

const Medium = 0.3
const High = 0.7

type VolumeAdjust struct {
	mutex      *sync.Mutex
	loudness   []float64 // [0 .. 1]
	colours    []colorful.Color
	timeout    bool
	reactivate time.Time
}

func (e *VolumeAdjust) Apply(sys *internal.System, t time.Time) {
	for _, led := range sys.LEDs {
		distance := e.distance(led.X, led.Y)
		intpart, decimal := math.Modf(distance * Intervals)
		distance_index := int(intpart)

		if distance_index >= Intervals {
			distance_index = Intervals - 1
		}

		e.mutex.Lock()
		next_value := e.loudness[distance_index+1]
		curr_value := e.loudness[distance_index]
		next_colour := e.colours[distance_index+1]
		curr_colour := e.colours[distance_index]
		e.mutex.Unlock()

		difference := next_value - curr_value
		adjusted := curr_value + difference*decimal

		if 0 <= adjusted && adjusted < Medium {
			v := uint8(float64(255) * adjusted)
			led.R, led.G, led.B = v, v, v
		} else if adjusted < High {
			v := curr_colour.BlendLuvLCh(next_colour, decimal)
			led.R, led.G, led.B = v.Clamped().RGB255()
		} else {
			// We set a one-second timeout if we go over our
			// loudness threshold
			e.mutex.Lock()
			if !e.timeout {
				e.timeout = true
				e.reactivate = time.Now().Add(time.Second * Timeout)
			}
			e.mutex.Unlock()
		}
	}
}

func (e *VolumeAdjust) distance(x, y float64) float64 {
	return y
}

func NewVolumeAdjust(deviceName string, debug bool) *VolumeAdjust {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	empty_loudness := make([]float64, 0, Intervals+1)
	for i := 0; i < Intervals+1; i++ {
		empty_loudness = append(empty_loudness, 0)
	}

	empty_colours := make([]colorful.Color, 0, Intervals+1)
	for i := 0; i < Intervals+1; i++ {
		empty_colours = append(empty_colours, colorful.LuvLCh(0.5, 0, 0))
	}

	effect := &VolumeAdjust{
		mutex:      new(sync.Mutex),
		loudness:   empty_loudness,
		colours:    empty_colours,
		timeout:    false,
		reactivate: time.Now(),
	}

	// Goroutine that runs in the background, constantly
	// picking up samples that are 1/20th of a second long
	go func() {
		samplesToRead := stream.SampleRate() / TimesPerSecond
		buf := make([]int16, samplesToRead)

		for {
			// wait/read until we have 1/20th of a second worth of audio
			err := readFull(stream, buf)
			if err != nil {
				panic(err)
			}

			// get the maximum amplitude here
			var max float64
			for _, v := range buf {
				if float64(v) > max {
					max = float64(v) / 4096.0
				} else if float64(-v) > max {
					max = float64(-v) / 4096.0
				}
			}

			buf64 := make([]float64, samplesToRead)

			for i := 0; i < samplesToRead; i++ {
				buf64[i] = float64(buf[i])
			}

			// fetch FFT results for the snippet (48k / 20 = 2.4k samples)
			fft_results := fft.FFTReal(buf64)
			colour := generateColour(fft_results, Threshold)

			current_time := time.Now()

			effect.mutex.Lock()
			if current_time.Before(effect.reactivate) {
				effect.loudness = appendRight(effect.loudness, 0)
				effect.colours = appendColour(effect.colours, colorful.FastLinearRgb(0.1, 0.1, 0.1))
			} else {
				if effect.timeout {
					effect.timeout = false
				}

				effect.loudness = appendRight(effect.loudness, max)
				effect.colours = appendColour(effect.colours, colour)
			}
			effect.mutex.Unlock()
		}
	}()

	return effect
}
