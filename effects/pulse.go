package effects

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"ledsim"

	"github.com/fogleman/ease"
	"github.com/google/uuid"
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

func PulseGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	totalTime := fadeIn + effect + fadeOut
	// target each fade to be about 15 seconds
	repeats := math.Round(float64(totalTime) / float64(15 * time.Second))
	playTime := time.Duration(float64(totalTime) / repeats)

	var keyframes []*ledsim.Keyframe

	keyframes = append(keyframes,
		&ledsim.Keyframe{
			Label:    "Pulse_FadeIn_" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn,
			Effect:   NewFadeTransition(FADE_IN),
			Layer:    2,
		},
		&ledsim.Keyframe{
			Label:    "Pulse_FadeOut_" + uuid.New().String(),
			Offset:   fadeIn + effect,
			Duration: fadeOut,
			Effect:   NewFadeTransition(FADE_OUT),
			Layer:    2,
		},
	)

	for i := 0; i < int(repeats); i++ {
		col := Golds[rng.Intn(len(Golds))]

		keyframes = append(keyframes,
			&ledsim.Keyframe{
				Label:    "Pulse_Main_" + strconv.Itoa(i) + "_" + uuid.New().String(),
				Offset:   time.Duration(i) * playTime,
				Duration: playTime,
				Effect: &Pulse{
					Dur:         playTime,
					BaseBright:  0.2,
					MaxBright:   0.8,
					TargetColor: col,
					LoDur:       time.Duration(float64(playTime) * 0.25),
					HiDur:       time.Duration(float64(playTime) * 0.25),
					UpDur:       time.Duration(float64(playTime) * 0.25),
					DownDur:     time.Duration(float64(playTime) * 0.25),
					EaseFunc:    ease.InCubic,
				},
				Layer: 1,
			},
		)
	}

	return keyframes
}
