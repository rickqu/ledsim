package main

import (
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type DiagonalRainbow struct {
}

func (d *DiagonalRainbow) Apply(sys *System, t time.Time) {
	secs := float64(t.UnixNano()) / float64(time.Second) // time.Second represents 1 second in nanoseconds

	for _, led := range sys.LEDs {
		v := math.Mod((secs+d.diagonalTravel(led.X+led.Z, led.Y+led.Z))*10, 360.0)
		led.R, led.G, led.B = colorful.Hcl(v, 1, 0.125).Clamped().RGB255()
	}
}

func (d *DiagonalRainbow) diagonalTravel(x, y float64) float64 {
	return x*math.Sin(0.785398) + y*math.Cos(0.785398)
}
