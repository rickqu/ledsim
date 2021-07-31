package effects

import (
	"fmt"
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

// type Bead struct {
// 	Position *ledsim.LED
// }

type FallingBeads struct {
	Beads   []*Inertia
	Visited map[*ledsim.LED]bool
}

// type Inertia struct {
// 	LED        *ledsim.LED
// 	ForwardLED *ledsim.LED
// 	Velocity   float64 // in units per second
// 	Progress   float64
// 	Fluid      bool
// 	Gravity    float64
// }

// speed is in LEDs per second
func NewFallingBeads() *FallingBeads {
	return &FallingBeads{
		Visited: make(map[*ledsim.LED]bool),
	}
}

func (b *FallingBeads) OnEnter(sys *ledsim.System) {
	fmt.Println("on enter")
	start := sys.DebugGetLEDByCoord(0.5, 0.5, 1)
	b.Visited[start] = true
	b.Visited[start.Neighbours[1]] = true
	b.Beads = []*Inertia{
		{
			LED:        start,
			ForwardLED: start.Neighbours[1],
			Velocity:   0.1,
			Progress:   0,
			Fluid:      true,
			Gravity:    0.1,
			Resistance: 5.0,
			Visited:    b.Visited,
		},
	}
}

func (b *FallingBeads) OnExit(sys *ledsim.System) {
}

func (b *FallingBeads) Eval(progress float64, sys *ledsim.System) {
	outBeads := make([]*Inertia, 0, len(b.Beads))
	for _, bead := range b.Beads {
		outBeads = append(outBeads, bead.Evaluate(progress)...)
	}

	b.Beads = outBeads

	ignoredBeads := make(map[*ledsim.LED]bool)
	for _, bead := range b.Beads {
		if bead.Velocity <= 0 {
			ignoredBeads[bead.LED] = true
		}
	}

	for led := range b.Visited {
		if ignoredBeads[led] {
			continue
		}

		led.Color = colorful.Color{0, 1, 0}
	}
}

var _ ledsim.Effect = (*FallingBeads)(nil)
