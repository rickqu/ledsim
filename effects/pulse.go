package effects

import (
	"ledsim"
	"math"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

type Pulse struct {
	Dur         time.Duration
	BaseBright  float64
	MaxBright   float64
	TargetColor colorful.Color
	LoDur       time.Duration
	HiDur       time.Duration
	UpDur       time.Duration
	DownDur     time.Duration
	EaseFunc    func(x float64) float64
}

func (p *Pulse) OnEnter(sys *ledsim.System) {

}

func (p *Pulse) OnExit(sys *ledsim.System) {

}

func (p *Pulse) Eval(progress float64, sys *ledsim.System) {
	var lumin float64

	leds := sys.LEDs

	total := p.LoDur + p.HiDur + p.UpDur + p.DownDur

	t := (float64(p.Dur) * progress / float64(time.Second))

	remainder := math.Mod(t, total.Seconds())

	if remainder < p.UpDur.Seconds() {
		lumin = p.EaseFunc(remainder / p.UpDur.Seconds())
	} else if remainder < (p.UpDur + p.HiDur).Seconds() {
		lumin = 1
	} else if remainder < (p.UpDur + p.HiDur + p.LoDur).Seconds() {
		lumin = 1 - p.EaseFunc((remainder-(p.UpDur+p.HiDur).Seconds())/p.LoDur.Seconds())
	} else {
		lumin = 0
	}

	lumin = lumin*(1.0-p.BaseBright) + p.BaseBright
	lumin *= p.MaxBright

	// d := (p.MaxBright-p.BaseBright)*(math.Sin(t*p.Speed*2.0*math.Pi)+1.0)/2.0 + p.BaseBright
	for _, led := range leds {

		led.Color = led.Color.BlendRgb(p.TargetColor, lumin)

		// led.Color.BlendRgb(colorful.LuvLCh(l, 1, v).Clamped(), l)
	}

	// fmt.Println(p.wave(t))

}
