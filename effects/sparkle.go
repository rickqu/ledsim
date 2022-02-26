package effects

import (
	"ledsim"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type Sparkle struct {
	ledPeriods []time.Duration
	delay      []time.Duration
	duration   time.Duration
	baseline   time.Duration
	deviation  time.Duration
	target     *colorful.Color
}

func NewSparkle(duration, baseline, deviation time.Duration, target *colorful.Color) *Sparkle {
	return &Sparkle{
		duration:  duration,
		baseline:  baseline,
		deviation: deviation,
		target:    target,
	}
}

func (s *Sparkle) OnEnter(sys *ledsim.System) {
	s.ledPeriods = make([]time.Duration, len(sys.LEDs))
	s.delay = make([]time.Duration, len(sys.LEDs))
	for i := range sys.LEDs {
		delta := time.Duration(rand.Float64() * float64(s.deviation))
		s.ledPeriods[i] = s.baseline + delta - (s.deviation / 2)
		s.delay[i] = time.Duration(rand.Float64() * float64(s.duration))
	}
}

func (s *Sparkle) Eval(progress float64, sys *ledsim.System) {
	t := time.Duration(progress * float64(s.duration))

	// each LED period is composed of 4 phases.
	for i, led := range sys.LEDs {
		t := t - s.delay[i]
		if t < 0 {
			continue
		}
		// determine our current block number
		block := t / (s.ledPeriods[i] * 4)
		totalBlocks := (s.duration - s.delay[i]) / (s.ledPeriods[i] * 4)
		locationInBlock := t % (s.ledPeriods[i] * 4)
		phase := locationInBlock / s.ledPeriods[i]
		locationInPhase := locationInBlock % s.ledPeriods[i]

		// do not turn on LEDs that are in their final block.
		if block >= totalBlocks {
			continue
		}

		switch phase {
		case 0:
			// black, do nothing
			break
		case 1:
			// fade in
			led.Color = ledsim.BlendRgb(led.Color, *s.target, float64(locationInPhase)/float64(s.ledPeriods[i]))
		case 2:
			// stay on
			led.Color = *s.target
		case 3:
			// fade out
			led.Color = ledsim.BlendRgb(led.Color, *s.target, 1-(float64(locationInPhase)/float64(s.ledPeriods[i])))
		}
	}
}

func (s *Sparkle) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*Sparkle)(nil)
