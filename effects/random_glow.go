package effects

import (
	"ledsim"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type RandomGlow struct {
	order     []int
	glow_time []float64
	real_dur  float64
	duration  time.Duration
	baseline  time.Duration
	deviation time.Duration
	start     colorful.Color
	end       colorful.Color
}

func NewRandomGlow(duration, baseline, deviation time.Duration, start, end colorful.Color) *RandomGlow {
	return &RandomGlow{
		order:     make([]int, TOTAL_LEDS),
		glow_time: make([]float64, TOTAL_LEDS),
		real_dur:  float64(duration),
		duration:  duration,
		baseline:  baseline,
		deviation: deviation,
		start:     start,
		end:       end,
	}
}

func (rg *RandomGlow) OnEnter(sys *ledsim.System) {
	rg.order = rand.Perm(TOTAL_LEDS)

	for i := TOTAL_LEDS - 1; i >= 0; i-- {
		glow_deviation := (rand.Float64() - 0.5) * float64(rg.deviation)
		glow_time := float64(rg.baseline) + glow_deviation

		led_start := rg.real_dur * (float64(i) / float64(TOTAL_LEDS))

		if led_start+glow_time > float64(rg.duration) {
			new_start := float64(rg.duration) - glow_time
			rg.real_dur = new_start * (float64(TOTAL_LEDS) / float64(i))
		}

		rg.glow_time[i] = glow_time
	}
}

func (rg *RandomGlow) Eval(progress float64, sys *ledsim.System) {
	current_time := progress * float64(rg.duration)

	for index, led_n := range rg.order {
		start_time := float64(rg.real_dur) * (float64(index) / float64(TOTAL_LEDS))
		current_glow := rg.glow_time[led_n]

		if start_time > current_time {
			sys.LEDs[led_n].Color = rg.start
			continue
		} else if start_time+current_glow < current_time {
			sys.LEDs[led_n].Color = rg.end
			continue
		}

		led_progress := (current_time - start_time) / current_glow
		sys.LEDs[led_n].Color = ledsim.BlendRgb(rg.start, rg.end, led_progress)
	}
}

func (rg *RandomGlow) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*RandomGlow)(nil)
