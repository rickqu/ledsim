package effects

import (
	"ledsim"
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type fader struct {
	dur time.Duration
	// dir	int
}

func NewFader(dur time.Duration) *fader {
	return &fader{
		dur: dur,
	}
}

func (p *fader) OnEnter(sys *ledsim.System) {

}

func (p *fader) OnExit(sys *ledsim.System) {

}

// func (p *fader) populate(sys *ledsim.System) {

// }

func (p *fader) Eval(progress float64, sys *ledsim.System) {
	leds := sys.LEDs

	// for _, led := range leds {
	// 	led.Color = colorful.Color{14, 0, 244}
	// }

	// for _, led := range leds {
	// 	if led.ID%2 == 1 {

	// 	}
	// 	led.B = colorful.FastHappyColor().B
	// }

	t := float64(p.dur) * progress / float64(time.Second) * 0.5

	//secs := float64(t.UnixNano()) / float64(time.Second)

	for _, led := range leds {
		v := math.Mod((t+p.diagonalTravel(led.X+led.Z, led.Y+led.Z))*50, 60)
		if v < 30 {
			// v = math.Mod(v, 30) + 20
			v += 30
		}

		//fmt.Println(v)
		led.R = colorful.LuvLCh(0.5, 1, v).Clamped().R
		led.G = colorful.LuvLCh(0.5, 1, v).Clamped().G
		led.B = colorful.LuvLCh(0.5, 1, v).Clamped().B
	}

}
func (p *fader) diagonalTravel(x, y float64) float64 {
	return x*math.Sin(1.57079) - y*math.Cos(0.785375)
}
