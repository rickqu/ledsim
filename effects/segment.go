package effects

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"ledsim"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
)

type Segment struct {
	*colorful.Color
	FADE_TYPE FADE_TYPE

	initialised bool
	chainToLeds map[int]*chainProgress
	chainOrder  []int
}

type chainProgress struct {
	leds     []*ledsim.LED
	progress float64
}

func newChainProgress() *chainProgress {
	return &chainProgress{make([]*ledsim.LED, 0), 0.0}
}

func (c *chainProgress) isDone() bool {
	return c.progress >= 1.0
}

func (c *chainProgress) appendLED(led *ledsim.LED) {
	c.leds = append(c.leds, led)
}

func NewSegment(c *colorful.Color, fadeType FADE_TYPE) *Segment {
	return &Segment{c, fadeType, false, make(map[int]*chainProgress), make([]int, 0)}
}

func (s *Segment) OnEnter(sys *ledsim.System) {
	if !s.initialised {
		for _, led := range sys.LEDs {
			_, chainExists := s.chainToLeds[led.Chain]
			if !chainExists {
				s.chainToLeds[led.Chain] = newChainProgress()
				s.chainOrder = append(s.chainOrder, led.Chain)
			}
			s.chainToLeds[led.Chain].appendLED(led)
		}
		s.initialised = true
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(s.chainOrder), func(i, j int) {
		s.chainOrder[i], s.chainOrder[j] = s.chainOrder[j], s.chainOrder[i]
	})
}

func (s *Segment) Eval(progress float64, sys *ledsim.System) {
	// if we are fading out, then we need to paint canvas with inital colour.
	var initialColour colorful.Color
	if s.FADE_TYPE == FADE_OUT {
		initialColour = *s.Color
	}
	for _, led := range sys.LEDs {
		led.Color = initialColour
	}

	var nthChain int
	var foundChain bool
	for i, chain := range s.chainOrder {
		if !s.chainToLeds[chain].isDone() {
			nthChain = i
			foundChain = true
			break
		} else {
			// chain is done, set it to full colour
			for _, led := range s.chainToLeds[chain].leds {
				if s.FADE_TYPE == FADE_IN {
					led.Color = *s.Color
				} else {
					led.Color = colorful.Color{0, 0, 0}
				}
			}
		}
	}
	if !foundChain {
		log.Println("warn: in segment animation all chains are done but animation is still going")
		return
	}

	stepSize := 1.0 / float64(len(s.chainOrder))
	startMilestone := float64(nthChain) * stepSize

	// If progress factor goes beyond 1 or under 0, then we can get random flashes of colour at the end of a fade.
	progressFactor := (progress - startMilestone) / stepSize
	if progressFactor > 1 {
		progressFactor = 1
	} else if progressFactor < 0 {
		progressFactor = 0
	}

	// fmt.Printf("Progress factor %f\n", progressFactor)
	var colourIntensity float64
	if s.FADE_TYPE == FADE_IN {
		colourIntensity = progressFactor
	} else {
		colourIntensity = 1 - progressFactor
	}

	for _, led := range s.chainToLeds[s.chainOrder[nthChain]].leds {
		led.Color.R = s.Color.R * colourIntensity
		led.Color.G = s.Color.G * colourIntensity
		led.Color.B = s.Color.B * colourIntensity
	}
	s.chainToLeds[s.chainOrder[nthChain]].progress = progressFactor
}

func (s *Segment) OnExit(sys *ledsim.System) {
	for _, chain := range s.chainToLeds {
		chain.progress = 0
	}
}

func SegmentGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	totalTime := fadeIn + effect + fadeOut
	repeats := math.Round(float64(totalTime) / float64(StandardPeriod))
	playTime := time.Duration(float64(totalTime) / repeats)

	var keyframes []*ledsim.Keyframe

	for i := 0; i < int(repeats); i++ {
		col := Golds[rng.Intn(len(Golds))]

		keyframes = append(keyframes,
			&ledsim.Keyframe{
				Label:    "Segment_MainIn_" + strconv.Itoa(i) + "_" + uuid.New().String(),
				Offset:   time.Duration(i) * playTime,
				Duration: playTime / 2,
				Effect:   NewSegment(&col, FADE_IN),
			},
			&ledsim.Keyframe{
				Label:    "Segment_MainOut_" + strconv.Itoa(i) + "_" + uuid.New().String(),
				Offset:   time.Duration(i)*playTime + (playTime / 2),
				Duration: playTime / 2,
				Effect:   NewSegment(&col, FADE_OUT),
			},
		)
	}

	return keyframes
}
