package effects

import (
	"math/rand"
	"time"

	"ledsim"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
)

type Segment struct {
	duration  time.Duration
	baseline  time.Duration
	deviation time.Duration
	palette   func() colorful.Color

	initialised bool
	chainToLeds map[int]*chain
	chainOrder  []int
}

type chain struct {
	leds   []*ledsim.LED
	period time.Duration
	delay  time.Duration
	colour colorful.Color
}

func NewSegment(duration, baseline, deviation time.Duration, palette func() colorful.Color) *Segment {
	return &Segment{
		duration:  duration,
		baseline:  baseline,
		deviation: deviation,
		palette:   palette,

		initialised: false,
		chainToLeds: make(map[int]*chain),
		chainOrder:  make([]int, 0),
	}
}

func (s *Segment) OnEnter(sys *ledsim.System) {
	if !s.initialised {
		for _, led := range sys.LEDs {
			_, chainExists := s.chainToLeds[led.Chain]
			if !chainExists {
				s.chainToLeds[led.Chain] = &chain{make([]*ledsim.LED, 0), 0, 0, colorful.Color{}}
				s.chainOrder = append(s.chainOrder, led.Chain)
			}
			s.chainToLeds[led.Chain].leds = append(s.chainToLeds[led.Chain].leds, led)
		}
		s.initialised = true
	}

	for _, chainToLed := range s.chainToLeds {
		delta := time.Duration(rand.Float64() * float64(s.deviation))
		chainToLed.period = s.baseline + delta - (s.deviation / 2)
		chainToLed.delay = time.Duration(rand.Float64() * float64(s.duration))
		chainToLed.colour = s.palette()
	}
}

func (s *Segment) Eval(progress float64, sys *ledsim.System) {
	t := time.Duration(progress * float64(s.duration))

	// each segment is composed of 4 phases.
	for _, chain := range s.chainToLeds {
		leds := chain.leds
		t := t - chain.delay
		if t < 0 {
			continue
		}
		// determine our current block number
		block := t / (chain.period * 4)
		totalBlocks := (s.duration - chain.delay) / (chain.period * 4)
		locationInBlock := t % (chain.period * 4)
		phase := locationInBlock / chain.period
		locationInPhase := locationInBlock % chain.period

		// do not turn on LEDs that are in their final block.
		if block >= totalBlocks {
			continue
		}

		for _, led := range leds {
			switch phase {
			case 0:
				// black, do nothing
				break
			case 1:
				// fade in
				led.Color = ledsim.BlendRgb(led.Color, chain.colour, float64(locationInPhase)/float64(chain.period))
			case 2:
				// stay on
				led.Color = chain.colour
			case 3:
				// fade out
				led.Color = ledsim.BlendRgb(led.Color, chain.colour, 1-(float64(locationInPhase)/float64(chain.period)))
			}
		}
	}
}

func (s *Segment) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*Segment)(nil)

func SegmentGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	gold := Golds[rand.Intn(len(Golds))]
	return []*ledsim.Keyframe{
		{
			Label:    "Segment_Main_" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn + fadeOut + effect,
			Effect: NewSegment(fadeIn+fadeOut+effect, 1*time.Second, 750*time.Millisecond,
				func() colorful.Color {
					return gold
				}),
			Layer: 1,
		},
	}
}

// baseline, deviation time.Duration, palette colorful.Color)
