package effects

import (
	"ledsim"
	"math"
	"sort"
	"time"

	"math/rand"

	"github.com/lucasb-eyer/go-colorful"
)

type AvoidingSnakeInstance struct {
	comps   []*ledsim.LED
	color   colorful.Color
	dur     time.Duration
	speed   float64
	curMove int
	head    int
}

type AvoidingSnake struct {
	snakes []*AvoidingSnakeInstance
}

type AvoidingSnakeConfig struct {
	Duration        time.Duration
	Speed           float64
	Palette         []colorful.Color
	RandomizeColors bool
	Head            int
	NumSnakes       int
	SnakeLength     int
}

type ScoringMap struct {
	scores map[*ledsim.LED]int
}

func (m *ScoringMap) GetScore(led *ledsim.LED) int {
	return m.scores[led]
}

func (a *AvoidingSnake) ComputeScoringMap(depth int) *ScoringMap {
	m := &ScoringMap{
		scores: make(map[*ledsim.LED]int),
	}

	type queueEntry struct {
		depth   int
		prevLED *ledsim.LED
		curLED  *ledsim.LED
	}

	for _, snake := range a.snakes {
		if snake.comps[0] == nil {
			continue
		}

		// flood fill from LED
		seen := make(map[*ledsim.LED]bool)
		queue := []queueEntry{
			{
				depth:   0,
				curLED:  snake.comps[len(snake.comps)-1],
				prevLED: snake.comps[len(snake.comps)-2],
			},
		}
		seen[snake.comps[len(snake.comps)-1]] = true
		seen[snake.comps[len(snake.comps)-2]] = true
		for len(queue) > 0 {
			top := queue[0]
			queue = queue[1:]

			for _, neigh := range top.curLED.Neighbours {
				if !seen[neigh] && neigh != top.prevLED {
					if top.depth >= depth {
						continue
					}

					seen[neigh] = true
					m.scores[neigh] += depth - top.depth
					queue = append(queue, queueEntry{
						depth:   top.depth + 1,
						curLED:  neigh,
						prevLED: top.curLED,
					})
				}
			}
		}
	}

	return m
}

func NewAvoidingSnake(config *AvoidingSnakeConfig) *AvoidingSnake {
	snake := &AvoidingSnake{
		snakes: make([]*AvoidingSnakeInstance, config.NumSnakes),
	}

	for i := range snake.snakes {
		snek := &AvoidingSnakeInstance{
			comps:   make([]*ledsim.LED, config.SnakeLength),
			dur:     config.Duration,
			speed:   config.Speed,
			curMove: 0,
			head:    config.Head,
		}

		if config.RandomizeColors {
			snek.color = config.Palette[rand.Intn(len(config.Palette))]
		} else {
			snek.color = config.Palette[i%len(config.Palette)]
		}

		snake.snakes[i] = snek
	}

	return snake
}

func (s *AvoidingSnake) OnEnter(sys *ledsim.System) {
	for _, snake := range s.snakes {
		m := s.ComputeScoringMap(100)
	candidateSearch:
		for {
			candidate := sys.LEDs[rand.Intn(len(sys.LEDs))]

			if m.GetScore(candidate) > 10 {
				continue
			}

			snake.comps[len(snake.comps)-1] = candidate
			for i := len(snake.comps) - 2; i >= 0; i-- {
				if len(snake.comps[i+1].Neighbours) <= 1 {
					continue candidateSearch
				}

				for _, next := range snake.comps[i+1].Neighbours {
					if next != snake.comps[i+1] &&
						(i+2 >= len(snake.comps) || next != snake.comps[i+2]) &&
						(i+3 >= len(snake.comps) || next != snake.comps[i+3]) {
						snake.comps[i] = next
					}
				}

				if snake.comps[i] == nil {
					continue candidateSearch
				}
			}

			break
		}
	}
}

// func (s *AvoidingSnake) populate(sys *ledsim.System, current *ledsim.LED, pos int, from *ledsim.LED) bool {
// 	for _, near := range current.Neighbours {
// 		if near == from {
// 			continue
// 		}

// 		s.snake[pos] = near
// 		if pos+1 < len(s.snake) {
// 			if !s.populate(sys, near, pos+1, current) {
// 				continue
// 			}
// 		}

// 		return true
// 	}

// 	return false
// }

func (m *ScoringMap) ScorePath(secondLast, last, to *ledsim.LED) int {
	// score the next 10 LEDs
	var score int
	curr := to
	for i := 0; i < 50; i++ {
		found := false
		for _, next := range to.Neighbours {
			if next == last || next == secondLast || curr == next {
				continue
			}

			score += m.GetScore(next)
			found = true
			secondLast, last, curr = last, curr, next
			break
		}

		if !found {
			return 1000000
		}
	}

	return score
}

func (a *AvoidingSnakeInstance) step(sys *ledsim.System, m *ScoringMap) bool {
	current := a.comps[len(a.comps)-1]
	sort.Slice(current.Neighbours, func(i, j int) bool {
		return m.ScorePath(a.comps[len(a.comps)-2], current, current.Neighbours[i])+rand.Intn(10) <
			m.ScorePath(a.comps[len(a.comps)-2], current, current.Neighbours[j])+rand.Intn(10)
	})

	for _, next := range current.Neighbours {
		if next == current || next == a.comps[len(a.comps)-2] ||
			next == a.comps[len(a.comps)-3] || next == a.comps[len(a.comps)-4] {
			continue
		}

		a.comps = append(a.comps[1:], next)
		return true
	}

	return false
}

func (s *AvoidingSnake) Eval(progress float64, sys *ledsim.System) {
	m := s.ComputeScoringMap(200)
	for _, snake := range s.snakes {
		snake.eval(progress, sys, m)
	}
}

func (a *AvoidingSnakeInstance) eval(progress float64, sys *ledsim.System, m *ScoringMap) {
	movement := ((progress * float64(a.dur)) / float64(time.Second)) * a.speed

	// move the snake
	intMov, frac := math.Modf(movement)
	for i := 0; i < int(intMov)-a.curMove; i++ {
		if !a.step(sys, m) {
			// reverse direction yolo
			for i, j := 0, len(a.comps)-1; i < j; i, j = i+1, j-1 {
				a.comps[i], a.comps[j] = a.comps[j], a.comps[i]
			}
			a.step(sys, m)
		}
	}

	a.curMove = int(intMov)

	for i := len(a.comps) - a.head; i >= 0; i-- {
		led := a.comps[i]
		if i == len(a.comps)-a.head {
			// the head
			led.Color = ledsim.BlendRgb(led.Color, a.color, (frac / 2))
			continue
		} else if i == 0 {
			// the tail
			led.Color = ledsim.BlendRgb(led.Color, a.color, 0.5-(frac/2))
			continue
		}

		led.Color = ledsim.BlendRgb(led.Color, a.color, 0.5)
	}
}

func (s *AvoidingSnake) OnExit(sys *ledsim.System) {

}

var _ ledsim.Effect = (*AvoidingSnake)(nil)
