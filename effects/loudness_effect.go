package effects

import (
	"fmt"
	"ledsim/internal"
	"math"
	"sync"
	"time"

	"github.com/1lann/dissonance/ffmpeg"
	"github.com/lucasb-eyer/go-colorful"
)

const Values = 40

type LoudnessEffect struct {
	mutex    *sync.Mutex
	loudness []float64 // [0 .. 1]
}

func (e *LoudnessEffect) Apply(sys *internal.System, t time.Time) {
	for _, led := range sys.LEDs {
		distance := e.distance(led.X, led.Y)
		intpart, diff_multiplier := math.Modf(distance * Values)
		distance_index := int(intpart)

		if distance_index >= Values {
			distance_index = Values - 1
		}

		e.mutex.Lock()
		next_value := e.loudness[(distance_index+1)%Values]
		curr_value := e.loudness[distance_index]
		e.mutex.Unlock()

		difference := next_value - curr_value

		v := 1 - (curr_value + difference*diff_multiplier)

		if v < 0 {
			v = 0
		}

		v = math.Pow(v, 15)

		col := colorful.Color{
			R: (float64(led.R) / 255.0) * v,
			G: (float64(led.G) / 255.0) * v,
			B: (float64(led.B) / 255.0) * v,
		}

		led.R, led.G, led.B = col.Clamped().RGB255()
	}
}

func (e *LoudnessEffect) distance(x, y float64) float64 {
	return (2.0 - math.Sqrt(x*x+y*y)) / 2
}

func appendRight(arr []float64, value float64) []float64 {
	rest := arr[1:]
	rest = append(rest, value)
	return rest
}

func NewLoudnessEffect(deviceName string, debug bool) *LoudnessEffect {
	stream, err := ffmpeg.NewFFMPEGStreamFromDshow(deviceName, debug)
	if err != nil {
		fmt.Println("tip: to enable ffmpeg debugging, pass \"debug\" as the third command argument")
		panic(err)
	}

	empty_loudness := make([]float64, 0, Values)
	for i := 0; i < Values; i++ {
		empty_loudness = append(empty_loudness, 0)
	}

	effect := &LoudnessEffect{
		mutex:    new(sync.Mutex),
		loudness: empty_loudness,
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

			// average the amplitude
			var max float64
			for _, v := range buf {
				if float64(v) > max {
					max = float64(v) / 4096.0
				} else if float64(-v) > max {
					max = float64(-v) / 4096.0
				}
			}

			// update the loudness
			effect.mutex.Lock()
			effect.loudness = appendRight(effect.loudness, max)
			effect.mutex.Unlock()
		}
	}()

	return effect
}
