package effects

import (
	"ledsim"
	"math"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

// A number coprime with the total number of LEDs (2162),
// useful in iterating through all LEDs without actually
// storing an array that's 2162 large
const INCREMENT = 45

// 689 * 91 = 1 (mod 2162)
const REVERSE = 1057

type Pseudorandom struct {
	initial  int
	duration time.Duration
	glow     time.Duration
	start    colorful.Color
	end      colorful.Color
}

// Create a new "random" effect that lights up random LEDs
// across the sculpture
func NewPseudorandom(duration, glow time.Duration, start, end colorful.Color) *Pseudorandom {
	return &Pseudorandom{
		initial:  rand.Intn(TOTAL_LEDS),
		duration: duration,
		glow:     glow,
		start:    start,
		end:      end,
	}
}

func (e *Pseudorandom) OnEnter(system *ledsim.System) {
	for _, led := range system.LEDs {
		led.Color = e.start
	}
}

func (e *Pseudorandom) Eval(progress float64, system *ledsim.System) {
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

	for index, led := range system.LEDs {
		reversed_index := index - e.initial

		if reversed_index < 0 {
			reversed_index += TOTAL_LEDS
		}

		reversed_index = (reversed_index * REVERSE) % TOTAL_LEDS

		if reversed_index < lower_bound {
			led.Color = e.end
			continue
		} else if reversed_index > upper_bound {
			led.Color = e.start
			continue
		}

		relative_index := upper_bound - reversed_index
		curr_frac := (fraction + float64(relative_index)) * real_time / (TOTAL_LEDS * float64(e.glow))
		led.Color = ledsim.BlendRgb(e.start, e.end, curr_frac)
	}
}

func (e *Pseudorandom) OnExit(system *ledsim.System) {
	for _, led := range system.LEDs {
		led.Color = e.end
	}
}

var _ ledsim.Effect = (*Pseudorandom)(nil)
