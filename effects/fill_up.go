package effects

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"ledsim"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
)

type FillUp struct {
	upDuration      time.Duration
	fadeOutDuration time.Duration
	fadeHeight      float64
	target          colorful.Color
	distFunc        func(led *ledsim.LED) float64
}

// NewFillUp creates an effect that fills lights from the bottom. It is assumed that the
// duration of the effect runs for upDuration + fadeOutDuration.
func NewFillUp(upDuration, fadeOutDuration time.Duration,
	fadeHeight float64, target colorful.Color, distFunc func(led *ledsim.LED) float64) *FillUp {
	return &FillUp{
		upDuration:      upDuration,
		fadeOutDuration: fadeOutDuration,
		fadeHeight:      fadeHeight,
		target:          target,
		distFunc:        distFunc,
	}
}

func (s *FillUp) OnEnter(sys *ledsim.System) {
}

func (s *FillUp) Eval(progress float64, sys *ledsim.System) {
	// log.Println("got progress:", progress)
	boundary := float64(s.upDuration) / float64(s.upDuration+s.fadeOutDuration)
	// log.Println("got boundary:", boundary)

	if progress < boundary {
		progress = progress / boundary

		// total distance to travel is 1 + s.fadeHeight
		fadeBoundary := progress/(1/(1+s.fadeHeight)) - s.fadeHeight

		for _, led := range sys.LEDs {
			darkness := (s.distFunc(led) - fadeBoundary) / s.fadeHeight
			if darkness < 0 {
				darkness = 0
			} else if darkness > 1 {
				darkness = 1
			}

			led.Color = ledsim.BlendRgb(led.Color, s.target, 1-darkness)
		}

		return
	}

	progress = (progress - boundary) / (1 - boundary)

	for _, led := range sys.LEDs {
		led.Color = ledsim.BlendRgb(led.Color, s.target, 1-progress)
	}
}

func (s *FillUp) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*FillUp)(nil)

var distFuncs = []func(led *ledsim.LED) float64{
	func(led *ledsim.LED) float64 {
		return led.Y
	},
	func(led *ledsim.LED) float64 {
		return led.Z
	},
	func(led *ledsim.LED) float64 {
		return led.X
	},
	func(led *ledsim.LED) float64 {
		return 1 - led.Y
	},
	func(led *ledsim.LED) float64 {
		return 1 - led.Z
	},
	func(led *ledsim.LED) float64 {
		return 1 - led.X
	},
	func(led *ledsim.LED) float64 {
		dist := math.Sqrt((led.X-0.5)*(led.X-0.5)+(led.Y-0.5)*(led.Y-0.5)+(led.Z-0.5)*(led.Z-0.5)) / 0.867
		if dist > 1 {
			dist = 1
		} else if dist < 0 {
			dist = 0
		}

		return dist
	},
	func(led *ledsim.LED) float64 {
		dist := math.Sqrt((led.X-0.5)*(led.X-0.5)+(led.Y-0.5)*(led.Y-0.5)+(led.Z-0.5)*(led.Z-0.5)) / 0.867
		if dist > 1 {
			dist = 1
		} else if dist < 0 {
			dist = 0
		}

		return dist
	},
}

func FillUpGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	totalTime := fadeIn + effect + fadeOut
	// target each fade to be about 7.5 seconds
	repeats := math.Round(float64(totalTime) / float64(StandardPeriod))
	playTime := time.Duration(float64(totalTime) / repeats)

	var keyframes []*ledsim.Keyframe

	keyframes = append(keyframes,
		&ledsim.Keyframe{
			Label:    "FillUp_FadeIn_" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn,
			Effect:   NewFadeTransition(FADE_IN),
			Layer:    2,
		},
		&ledsim.Keyframe{
			Label:    "FillUp_FadeOut_" + uuid.New().String(),
			Offset:   fadeIn + effect,
			Duration: fadeOut,
			Effect:   NewFadeTransition(FADE_OUT),
			Layer:    2,
		},
	)

	for i := 0; i < int(repeats); i++ {
		col := Golds[rng.Intn(len(Golds))]

		upDuration := time.Duration(float64(playTime) * 0.7)
		downDuration := time.Duration(float64(playTime) * 0.3)

		keyframes = append(keyframes,
			&ledsim.Keyframe{
				Label:    "FillUp_Main_" + strconv.Itoa(i) + "_" + uuid.New().String(),
				Offset:   time.Duration(i) * playTime,
				Duration: playTime,
				Effect:   NewFillUp(upDuration, downDuration, 0.2, col, distFuncs[rng.Intn(len(distFuncs))]),
				Layer:    1,
			},
		)
	}

	return keyframes
}
