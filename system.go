package ledsim

import (
	"sync"

	"github.com/lucasb-eyer/go-colorful"
)

type System struct {
	LEDs    []*LED
	Teensys map[string]*Teensy

	XStats        *Stats
	YStats        *Stats
	ZStats        *Stats
	normalizeOnce *sync.Once
}

type PhysicalLEDPosition struct {
	TeensyIp        string
	Chain           int
	PositionOnChain int
}

type RGBOutput struct {
	Red   *byte
	Green *byte
	Blue  *byte
}

type LED struct {
	ID int
	X  float64
	Y  float64
	Z  float64
	PhysicalLEDPosition
	RGBOutput
	RawLine string

	colorful.Color
	Neighbours []*LED
}

type Chain struct {
	Id       int
	Pin      int
	PosOnPin int
	Length   int
	Reversed bool
}

type Teensy struct {
	IP     string
	Chains map[int]*Chain
}

type Middleware interface {
	Execute(system *System, next func() error) error
}

type MiddlewareFunc func(system *System, next func() error) error

func (m MiddlewareFunc) Execute(system *System, next func() error) error {
	return m(system, next)
}

func NewSystem() *System {
	return &System{
		normalizeOnce: new(sync.Once),
	}
}

func (s *System) AddLED(led *LED) {
	led.ID = len(s.LEDs)
	s.LEDs = append(s.LEDs, led)
}

type Stats struct {
	Min float64
	Max float64
}

func (s *Stats) Convert(val float64) float64 {
	return (val - s.Min) / (s.Max - s.Min)
}

func (s *System) computeStats(getter func(led *LED) float64) *Stats {
	min := 100000000.0
	max := -100000000.0
	for _, led := range s.LEDs {
		v := getter(led)
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	return &Stats{
		Min: min,
		Max: max,
	}
}

func (s *System) Normalize() {
	for _, led := range s.LEDs {
		led.X = -led.X
	}

	s.normalizeOnce.Do(func() {
		s.XStats = s.computeStats(func(led *LED) float64 {
			return led.X
		})
		s.YStats = s.computeStats(func(led *LED) float64 {
			return led.Y
		})
		s.ZStats = s.computeStats(func(led *LED) float64 {
			return led.Z
		})

		for _, led := range s.LEDs {
			led.X = s.XStats.Convert(led.X)
			led.Y = s.YStats.Convert(led.Y)
			led.Z = s.ZStats.Convert(led.Z)
		}
	})
}

func (s *System) DebugGetLEDByCoord(x float64, y float64, z float64) *LED {
	minDist := 1.0
	var curVertex *LED
	for _, v := range s.LEDs {
		dist := getDistance(x, y, z, v.X, v.Y, v.Z)
		if dist < minDist {
			minDist = dist
			curVertex = v
		}
	}
	if curVertex == nil {
		panic("get vertex by coord failed")
	}
	return curVertex
}
