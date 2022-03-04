package effects

import (
	"math/rand"
	"time"

	"ledsim"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
)

type ColourShift struct {
	colours []colorful.Color
}

func NewColourShift(colours []colorful.Color) *ColourShift {
	return &ColourShift{
		colours: append(colours, append(colours, colours...)...),
	}
}

func (s *ColourShift) OnEnter(sys *ledsim.System) {

}

func (s *ColourShift) Eval(progress float64, sys *ledsim.System) {
	startColourIndex := int(progress * float64(len(s.colours)-1))

	for _, led := range sys.LEDs {
		led.Color = ledsim.BlendRgb(s.colours[startColourIndex], s.colours[startColourIndex+1], s.progressOnCurrentColour(startColourIndex, progress))
	}
}

func (s *ColourShift) progressOnCurrentColour(startColourIndex int, totalAnimationProgress float64) float64 {
	animationPercentagePerColourShift := 1.0 / float64(len(s.colours)-1)
	return (totalAnimationProgress - float64(startColourIndex)*animationPercentagePerColourShift) / animationPercentagePerColourShift
}

func (s *ColourShift) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*ColourShift)(nil)

func ColourShiftGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	return []*ledsim.Keyframe{
		{
			Label:    "ColourShift_FadeIn_" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn,
			Effect:   NewFadeTransition(FADE_IN),
			Layer:    2,
		},
		{
			Label:    "ColourShift_FadeIn_Background" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn,
			Effect:   NewMonocolour(Golds[0]),
			Layer:    1,
		},
		{
			Label:    "ColourShift_Main_" + uuid.New().String(),
			Offset:   fadeIn,
			Duration: effect,
			Effect:   NewColourShift(Golds),
			Layer:    1,
		},
		{
			Label:    "ColourShift_FadeOut_Background" + uuid.New().String(),
			Offset:   fadeIn + effect,
			Duration: fadeOut,
			Effect:   NewMonocolour(Golds[len(Golds)-1]),
			Layer:    1,
		},
		{
			Label:    "ColourShift_FadeOut_" + uuid.New().String(),
			Offset:   fadeIn + effect,
			Duration: fadeOut,
			Effect:   NewFadeTransition(FADE_OUT),
			Layer:    2,
		},
	}
}

// baseline, deviation time.Duration, palette colorful.Color)
