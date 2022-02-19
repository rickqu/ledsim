package effects

import (
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

type FADE_TYPE int

const (
	FADE_IN FADE_TYPE = iota
	FADE_OUT
)

type FadeTransition struct {
	fadeType FADE_TYPE
}

func NewFadeTransition(fadeType FADE_TYPE) *FadeTransition {
	return &FadeTransition{fadeType}
}

func (s *FadeTransition) OnEnter(sys *ledsim.System) {
}

func (s *FadeTransition) Eval(progress float64, sys *ledsim.System) {
	for _, led := range sys.LEDs {
		var brightnessMultiplier float64
		if s.fadeType == FADE_IN {
			brightnessMultiplier = progress
		} else {
			brightnessMultiplier = 1 - progress
		}
		led.Color = colorful.Color{led.Color.R * brightnessMultiplier, led.Color.G * brightnessMultiplier, led.Color.B * brightnessMultiplier}
	}
}

func (s *FadeTransition) OnExit(sys *ledsim.System) {
}
