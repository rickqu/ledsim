package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/1lann/dissonance/audio"
	"github.com/1lann/dissonance/ffmpeg"
)

type BasicMicEffect struct {
	mutex    *sync.Mutex
	loudness float64 // [0 .. 1]
}

func (e *BasicMicEffect) Apply(sys *System, t time.Time) {
	var loudness float64
	e.mutex.Lock()
	loudness = e.loudness
	e.mutex.Unlock()

	if loudness < 0.1 {
		loudness = 0.1
	}
	if loudness >= 0.98 {
		loudness = 0.98
	}

	loudness = (1 - loudness) / 2

	for _, led := range sys.LEDs {
		r := uint8(float64(led.R) * loudness)
		g := uint8(float64(led.G) * loudness)
		b := uint8(float64(led.B) * loudness)
		led.R, led.G, led.B = r, g, b
	}
}

func readFull(stream audio.Stream, buf []int16) error {
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

func NewBasicMicEffect(deviceName string, debug bool) *BasicMicEffect {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	effect := &BasicMicEffect{
		mutex: new(sync.Mutex),
	}

	go func() {
		samplesToRead := stream.SampleRate() / 20

		buf := make([]int16, samplesToRead)

		for {
			// wait/read until we have 1/20th of a second worth of audio
			err := readFull(stream, buf)
			if err != nil {
				panic(err)
			}

			// max the amplitude
			var max float64
			for _, v := range buf {
				if float64(v) > max {
					max = float64(v) / 4096.0
				} else if float64(-v) > max {
					max = float64(-v) / 4096.0
				}
			}

			max *= 2
			if max > 1 {
				max = 1
			}

			// update the loudness
			effect.mutex.Lock()
			if max < effect.loudness {
				effect.loudness -= 0.03
			} else {
				effect.loudness = max
			}
			effect.mutex.Unlock()
		}

	}()

	return effect
}
