package effects

import (
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

var count int = 0

type Monocolour struct {
	*colorful.Color
}

func NewMonocolour(Colour *colorful.Color) *Monocolour {
	return &Monocolour{Colour}
}

func (s *Monocolour) OnEnter(sys *ledsim.System) {
}

func (s *Monocolour) Eval(progress float64, sys *ledsim.System) {
	for _, led := range sys.LEDs {
		led.Color = *s.Color
	}
}

func (s *Monocolour) OnExit(sys *ledsim.System) {
}
