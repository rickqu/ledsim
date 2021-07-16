package effects

import (
	"ledsim"

	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
)

// type Bead struct {
// 	Position *ledsim.LED
// }

type FloodFill struct {
	start        *ledsim.LED
	distMap      map[*ledsim.LED]int
	maxGrowth    float64
	color        colorful.Color
	fadeOut      FadeOut
	fadeOutStart float64
}

// type Inertia struct {
// 	LED        *ledsim.LED
// 	ForwardLED *ledsim.LED
// 	Velocity   float64 // in units per second
// 	Progress   float64
// 	Fluid      bool
// 	Gravity    float64
// }

type FadeOut int

const (
	FadeOutRipple = iota
	FadeOutFade
)

// speed is in LEDs per second
func NewFloodFill(start *ledsim.LED, maxGrowth float64, color colorful.Color,
	fadeOut FadeOut, fadeOutStart float64) *FloodFill {
	return &FloodFill{
		start:        start,
		distMap:      make(map[*ledsim.LED]int),
		maxGrowth:    maxGrowth,
		color:        color,
		fadeOut:      fadeOut,
		fadeOutStart: fadeOutStart,
	}
}

func (b *FloodFill) OnEnter(sys *ledsim.System) {
	type queueEntry struct {
		depth int
		led   *ledsim.LED
	}

	queue := []queueEntry{
		{
			depth: 0,
			led:   b.start,
		},
	}

	b.distMap[b.start] = 0

	for len(queue) > 0 {
		top := queue[0]
		queue = queue[1:]

		for _, neighbor := range top.led.Neighbours {
			if _, found := b.distMap[neighbor]; found {
				continue
			}

			b.distMap[neighbor] = top.depth + 1
			queue = append(queue, queueEntry{
				depth: top.depth + 1,
				led:   neighbor,
			})
		}
	}
}

func (b *FloodFill) OnExit(sys *ledsim.System) {
}

func (b *FloodFill) Eval(progress float64, sys *ledsim.System) {
	width := (progress / 0.5) * b.maxGrowth // 0 to 1 passed to decayFunc is mapped to this range
	if progress > 0.5 {
		width = b.maxGrowth
	}

	for led, dist := range b.distMap {
		if float64(dist) > width {
			continue
		}

		prevColor := led.Color

		// apply decay only to last 10 LEDs
		if width-float64(dist) < 10 {
			led.Color = ledsim.BlendRgb(led.Color, b.color, ease.OutExpo((width-float64(dist))/10.0))
		} else {
			led.Color = ledsim.BlendRgb(led.Color, b.color, 1.0)
		}

		if progress > 0.5 {
			if b.fadeOut == FadeOutRipple {
				// reverse it a bit lol
				rippleWidth := ((progress - b.fadeOutStart) / (1 - b.fadeOutStart)) * b.maxGrowth
				if float64(dist) <= rippleWidth {

					if rippleWidth-float64(dist) < 10 {
						led.Color = ledsim.BlendRgb(led.Color, prevColor, ease.OutExpo((rippleWidth-float64(dist))/10.0))
					} else {
						led.Color = ledsim.BlendRgb(led.Color, prevColor, 1.0)
					}
				}
			} else if b.fadeOut == FadeOutFade {
				led.Color = ledsim.BlendRgb(led.Color, prevColor, (progress-b.fadeOutStart)/(1-b.fadeOutStart))
			}
		}
	}
}

var _ ledsim.Effect = (*FloodFill)(nil)
