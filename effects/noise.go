package effects

import (
	"ledsim"
	"math"
	"math/rand"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

const ARCS = 100

type Noise struct {
	iteration int
	previous  []float64
	next      []float64
	duration  time.Duration
	cycle     time.Duration
	height    float64
	deviation float64
	start     colorful.Color
	end       colorful.Color
}

// Create a new "Windows Media Player"-esque "noise" function.
// - `duration`: how long the animation lasts for
// - `cycle`: how often we want to update our "bars" - the shorter
//   the duration, the more jittery the animation becomes
// - `height`: where the "baseline" for the animation is relative to
//   the sculpture, so 0.5 means the animation is roughly halfway up
// - `deviation`: how high or low the bars are
func NewNoise(duration, cycle time.Duration, height, deviation float64, start, end colorful.Color) *Noise {
	return &Noise{
		iteration: 0,
		previous:  make([]float64, ARCS),
		next:      make([]float64, ARCS),
		duration:  duration,
		cycle:     cycle,
		height:    height,
		deviation: deviation,
		start:     start,
		end:       end,
	}
}

func (n *Noise) populate() {
	n.previous = n.next

	for i := 0; i < ARCS; i++ {
		n.next[i] = n.height + (rand.Float64()-0.5)*n.deviation
	}
}

func (n *Noise) OnEnter(sys *ledsim.System) {
	n.populate()
	n.previous = n.next
}

func (n *Noise) Eval(progress float64, sys *ledsim.System) {
	current_time := progress * float64(n.duration)
	intpart, fraction := math.Modf(current_time / float64(n.cycle))

	for int(intpart) > n.iteration {
		n.populate()
		n.iteration++
	}

	for _, led := range sys.LEDs {
		angle := math.Atan2(led.Y, led.X) + math.Pi
		section := int(angle / (2 * math.Pi))

		cutoff := n.previous[section] + fraction*(n.next[section]-n.previous[section])

		if led.Z > cutoff {
			led.Color = n.start
		} else {
			led.Color = n.end
		}
	}
}

func (n *Noise) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*Noise)(nil)
