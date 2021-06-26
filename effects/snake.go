package effects

import (
	"ledsim"
	"math"
	"time"

	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
)

type Snake struct {
	curMove int
	snake   []*ledsim.LED
	dur     time.Duration
	speed   float64
	col     colorful.Color
}

// speed is in LEDs per second
func NewSnake(dur time.Duration, speed float64, col colorful.Color) *Snake {
	return &Snake{
		snake: make([]*ledsim.LED, 50),
		dur:   dur,
		speed: speed,
		col:   col,
	}
}

func (s *Snake) OnEnter(sys *ledsim.System) {
	// pick a random LED to start from
	start := sys.LEDs[rand.Intn(len(sys.LEDs))]

	s.snake[0] = start
	s.populate(sys, start, 1, nil)
}

func (s *Snake) populate(sys *ledsim.System, current *ledsim.LED, pos int, from *ledsim.LED) bool {
	for _, near := range current.Neighbours {
		if near == from {
			continue
		}

		s.snake[pos] = near
		if pos+1 < len(s.snake) {
			if !s.populate(sys, near, pos+1, current) {
				continue
			}
		}

		return true
	}

	return false
}

func (s *Snake) step(sys *ledsim.System) bool {
	current := s.snake[len(s.snake)-1]
	for attempts := 0; attempts < 100; attempts++ {
		near := current.Neighbours[rand.Intn(len(current.Neighbours))]

		if near == s.snake[len(s.snake)-2] {
			continue
		}

		s.snake = append(s.snake[1:], near)
		return true
	}

	// fmt.Println("im at LED ID:", current.ID, "raw:", current.RawLine)
	// for _, neigh := range current.Neighbours {
	// 	fmt.Println("neighbours are:", neigh.ID, "raw:", neigh.RawLine)
	// }

	return false
}

func (s *Snake) Eval(progress float64, sys *ledsim.System) {
	movement := ((progress * float64(s.dur)) / float64(time.Second)) * s.speed

	// move the snake
	intMov, frac := math.Modf(movement)
	for i := 0; i < int(intMov)-s.curMove; i++ {
		if !s.step(sys) {
			// reverse direction yolo
			for i, j := 0, len(s.snake)-1; i < j; i, j = i+1, j-1 {
				s.snake[i], s.snake[j] = s.snake[j], s.snake[i]
			}
			s.step(sys)
		}
	}

	s.curMove = int(intMov)

	col := s.col

	for i, led := range s.snake {
		if i == 0 {
			// the tail
			led.Color = ledsim.BlendRgb(led.Color, col, 0.5-(frac/2))
			continue
		} else if i == len(s.snake)-1 {
			// the head
			led.Color = ledsim.BlendRgb(led.Color, col, (frac / 2))
			continue
		}

		led.Color = ledsim.BlendRgb(led.Color, col, 0.5)
	}
}

func (s *Snake) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*Snake)(nil)
