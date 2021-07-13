package effects

import (
	"ledsim"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type RandomGlow struct {
	order     []int
	glow_time []float64
	duration  time.Duration
	minimum   time.Duration
	maximum   time.Duration
	start     colorful.Color
	end       colorful.Color
}

func NewRandomGlow(duration, minimum, maximum time.Duration, start, end colorful.Color) *RandomGlow {
	return &RandomGlow{
		order:     make([]int, TOTAL_LEDS),
		glow_time: make([]float64, TOTAL_LEDS),
		duration:  duration,
		minimum:   minimum,
		maximum:   maximum,
		start:     start,
		end:       end,
	}
}

func (rg *RandomGlow) OnEnter(sys *ledsim.System) {

}

func (rg *RandomGlow) Eval(progress float64, sys *ledsim.System) {

}

func (rg *RandomGlow) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*RandomGlow)(nil)
