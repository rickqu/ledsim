package effects

import (
	"ledsim"
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type PanningFade struct {
	dur   time.Duration
	speed float64
	// dir	float64
}

func NewPanningFade(dur time.Duration, speed float64) *PanningFade {
	return &PanningFade{
		dur:   dur,
		speed: speed,
	}
}

func (p *PanningFade) OnEnter(sys *ledsim.System) {

}

func (p *PanningFade) OnExit(sys *ledsim.System) {

}

// func (p *PanningFade) populate(sys *ledsim.System) {

// }

func (p *PanningFade) Eval(progress float64, sys *ledsim.System) {
	leds := sys.LEDs

	// for _, led := range leds {
	// 	led.Color = colorful.Color{14, 0, 244}
	// }

	t := float64(p.dur) * progress / float64(time.Second) * float64(p.speed)

	for _, led := range leds {
		v := math.Mod((t+p.diagonalTravel(led.X+led.Z, led.Y+led.Z))*50, 100)
		l := 0.5
		if v < 30 {
			// v = math.Mod(v, 30) + 20
			v += 30
		}

		if v > 90 {
			l = (-1 / 20) * (v - 100)
		} else if v < 40 {
			l = 0.05 * (v - 30)
		}

		if l < 0 {
			l = 0
		} else if l > 1 {
			l = 0.5
		}

		//fmt.Println(v)
		led.R = colorful.LuvLCh(l, 1, v).Clamped().R
		led.G = colorful.LuvLCh(l, 1, v).Clamped().G
		led.B = colorful.LuvLCh(l, 1, v).Clamped().B
	}

}
func (p *PanningFade) diagonalTravel(x, y float64) float64 {
	return x*math.Sin(1.57079) + y*math.Cos(1.57079)
}
