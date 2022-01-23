package effects

import (
	"ledsim"
	"math"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type Random struct {
	order    []int
	duration time.Duration
	glow     time.Duration
	start    colorful.Color
	end      colorful.Color
}

func NewRandom(duration, glow time.Duration, start, end colorful.Color) *Random {
	return &Random{
		order:    make([]int, TOTAL_LEDS),
		duration: duration,
		glow:     glow,
		start:    start,
		end:      end,
	}
}

func (e *Random) OnEnter(system *ledsim.System) {
	e.order = rand.Perm(TOTAL_LEDS)
}

func (e *Random) Eval(progress float64, system *ledsim.System) {
	// our real animation time
	real_time := float64(e.duration - e.glow)

	// the current time in our animation
	current_time := progress * float64(e.duration)

	cycle_behind := current_time - float64(e.glow)
	intpart, fraction := math.Modf(cycle_behind * TOTAL_LEDS / real_time)

	// the least recent LED index we have to change
	lower_bound := int(intpart)

	// checks if the lower bound is negative, in which case
	// we set it to 0 and set our fraction to where it should be
	if math.Signbit(intpart) {
		lower_bound = 0
		fraction = 1 + fraction
	}

	// the most recent LED index we have to change
	upper_bound := int(math.Floor(current_time * TOTAL_LEDS / real_time))

	if upper_bound > TOTAL_LEDS {
		upper_bound = TOTAL_LEDS - 1
	}

	for index, led := range e.order {
		current_led := system.LEDs[led]

		if index < lower_bound {
			current_led.Color = e.end
			continue
		} else if index > upper_bound {
			current_led.Color = e.start
			continue
		}

		relative_index := upper_bound - index
		curr_frac := (fraction + float64(relative_index)) * real_time / (TOTAL_LEDS * float64(e.glow))
		current_led.Color = ledsim.BlendRgb(e.start, e.end, curr_frac)
	}
}

func (e *Random) OnExit(system *ledsim.System) {

}

var _ ledsim.Effect = (*Random)(nil)
