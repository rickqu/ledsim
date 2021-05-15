package main

import (
	"sync"
	"time"
)

type LED struct {
	X float64
	Y float64
	Z float64
	R uint8
	G uint8
	B uint8
}

func (l *LED) RGB() uint32 {
	return uint32(l.R)<<16 | uint32(l.G)<<8 | uint32(l.B)
}

type System struct {
	LEDs               []*LED
	afterFrameCallback func(s *System, t time.Time)

	XStats        *Stats
	YStats        *Stats
	ZStats        *Stats
	normalizeOnce *sync.Once
}

func NewSystem() *System {
	return &System{
		afterFrameCallback: func(s *System, t time.Time) {},
	}
}

func (s *System) AddLED(led *LED) {
	s.LEDs = append(s.LEDs, led)
}

func (s *System) AfterFrame(callback func(s *System, t time.Time)) {
	s.afterFrameCallback = callback
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

func (s *System) normalize() {
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

func (s *System) Run(effects []Effect) {
	s.normalize()
	// main system runtime (20 fps)
	ticker := time.NewTicker(time.Second / 20)
	for range ticker.C {
		now := time.Now()
		for _, effect := range effects {
			effect.Apply(s, now)
		}

		s.afterFrameCallback(s, now)
	}
}
