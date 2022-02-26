package effects

import (
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

type Segment struct {
	colorful.Color
}

func NewSegment(Colour colorful.Color) *Monocolour {
	return &Monocolour{Colour}
}

func (s *Segment) OnEnter(sys *ledsim.System) {
}

func (s *Segment) Eval(progress float64, sys *ledsim.System) {
	for _, led := range sys.LEDs {
		led.Color = s.Color
	}
}

func (s *Segment) OnExit(sys *ledsim.System) {
}
