package effects

import (
	"ledsim"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type FillUp struct {
	upDuration      time.Duration
	fadeOutDuration time.Duration
	fadeHeight      float64
	target          colorful.Color
}

// NewFillUp creates an effect that fills lights from the bottom. It is assumed that the
// duration of the effect runs for upDuration + fadeOutDuration.
func NewFillUp(upDuration, fadeOutDuration time.Duration,
	fadeHeight float64, target colorful.Color) *FillUp {
	return &FillUp{
		upDuration:      upDuration,
		fadeOutDuration: fadeOutDuration,
		fadeHeight:      fadeHeight,
		target:          target,
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
			darkness := (led.Z - fadeBoundary) / s.fadeHeight
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
