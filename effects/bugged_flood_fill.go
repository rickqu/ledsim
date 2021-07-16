package effects

import (
	"ledsim"

	"github.com/lucasb-eyer/go-colorful"
)

type BuggedFloodFill struct {
	start     *ledsim.LED
	distMap   map[*ledsim.LED]int
	maxGrowth float64
	decayFunc func(v float64) float64
	color     colorful.Color
	fadeOut   FadeOut
}

// speed is in LEDs per second
func NewBuggedFloodFill(start *ledsim.LED, maxGrowth float64, color colorful.Color,
	fadeOut FadeOut, decayFunc func(v float64) float64) *BuggedFloodFill {
	return &BuggedFloodFill{
		start:     start,
		distMap:   make(map[*ledsim.LED]int),
		maxGrowth: maxGrowth,
		decayFunc: decayFunc,
		color:     color,
		fadeOut:   fadeOut,
	}
}

func (b *BuggedFloodFill) OnEnter(sys *ledsim.System) {
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

func (b *BuggedFloodFill) OnExit(sys *ledsim.System) {
}

func (b *BuggedFloodFill) Eval(progress float64, sys *ledsim.System) {
	width := (progress / 0.5) * b.maxGrowth // 0 to 1 passed to decayFunc is mapped to this range
	if progress > 0.5 {
		width = b.maxGrowth
	}

	for led, dist := range b.distMap {
		if float64(dist) > width {
			continue
		}

		prevColor := led.Color
		led.Color = ledsim.BlendRgb(led.Color, b.color, b.decayFunc(float64(dist)/width))
		if b.fadeOut == FadeOutRipple {
			// reverse it a bit lol
			rippleWidth := (progress - 0.5) / 0.5 * b.maxGrowth
			led.Color = ledsim.BlendRgb(led.Color, prevColor, b.decayFunc(float64(dist)/rippleWidth))
		} else if b.fadeOut == FadeOutFade {
			led.Color = ledsim.BlendRgb(led.Color, prevColor, (progress-0.5)/0.5)
		}
	}
}

var _ ledsim.Effect = (*BuggedFloodFill)(nil)
