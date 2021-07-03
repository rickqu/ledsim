package effects

import (
	"ledsim"
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type SegmentShift struct {
	segments map[int][]*ledsim.LED
	dur      time.Duration
	speed    float64
	onWidth  int
	offWidth int
	color    colorful.Color
}

// speed is in LEDs per second
func NewSegmentShift(dur time.Duration, speed float64, onWidth int, offWidth int, col colorful.Color) *SegmentShift {
	return &SegmentShift{
		segments: make(map[int][]*ledsim.LED),
		dur:      dur,
		speed:    speed,
		onWidth:  onWidth,
		offWidth: offWidth,
		color:    col,
	}
}

func (s *SegmentShift) OnEnter(sys *ledsim.System) {
	s.populate(sys)
}

func (s *SegmentShift) OnExit(sys *ledsim.System) {
}

func (s *SegmentShift) populate(sys *ledsim.System) {
	type queueEntry struct {
		led     *ledsim.LED
		segment int
	}

	q := []queueEntry{
		{
			led:     sys.LEDs[0],
			segment: 0,
		},
	}

	seen := make(map[*ledsim.LED]bool)
	seen[sys.LEDs[0]] = true

	for len(q) > 0 {
		top := q[0]
		q = q[1:]

		s.segments[top.segment] = append(s.segments[top.segment], top.led)

		for _, led := range top.led.Neighbours {
			if seen[led] {
				continue
			}
			seen[led] = true
			q = append(q, queueEntry{
				led:     led,
				segment: (top.segment + 1) % (s.onWidth + s.offWidth),
			})
		}
	}
}

func (s *SegmentShift) Eval(progress float64, sys *ledsim.System) {
	movement := ((progress * float64(s.dur)) / float64(time.Second)) * s.speed

	// move the snake
	intMov, frac := math.Modf(movement)

	mov := int(intMov) % (s.onWidth + s.offWidth)

	for i := mov; i < mov+s.onWidth; i++ {
		a := i % (s.onWidth + s.offWidth)
		if i == mov {
			for _, led := range s.segments[a] {
				led.Color = ledsim.BlendRgb(led.Color, s.color, 0.5-(frac/2))
			}
		} else if i == mov+s.onWidth-1 {
			for _, led := range s.segments[a] {
				led.Color = ledsim.BlendRgb(led.Color, s.color, (frac / 2))
			}
		} else {
			for _, led := range s.segments[a] {
				led.Color = ledsim.BlendRgb(led.Color, s.color, 1.0)
			}
		}
	}
}

var _ ledsim.Effect = (*SegmentShift)(nil)
