package effects

import (
	"ledsim"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

func GetEffects() []*ledsim.Keyframe {
	golds := []colorful.Color{
		// {255, 255, 0},
		// {212, 175, 55},
		// {207, 181, 59},
		// {197, 179, 88},
		{230, 190, 138},
		// {153, 101, 21},
		// {244, 163, 0},
	}

	for i, gold := range golds {
		gold.R = gold.R / 255.0
		gold.G = gold.G / 255.0
		gold.B = gold.B / 255.0
		golds[i] = gold
		// fmt.Println(gold)
	}

	return []*ledsim.Keyframe{
		{
			Label:    "black background",
			Offset:   0,
			Duration: time.Hour,
			Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
				led.Color = colorful.Color{0, 0, 0}
			}),
			Layer: -1000,
		},
		// {
		// 	Label:    "shooting star test",
		// 	Offset:   0,
		// 	Duration: time.Second,
		// 	Effect:   effects.NewShootingStar(effects.Vector{0, 0, 0}, effects.Vector{1, 1, 1}),
		// },
		{
			Label:    "segment test",
			Offset:   0,
			Duration: time.Second * 30,
			Effect:   NewSegmentShift(time.Second*5, 50, 30, 70, golds[0]),
		},
		{
			Label:    "good snake settings",
			Offset:   time.Second * 30,
			Duration: time.Second * 600,
			Effect: NewAvoidingSnake(&AvoidingSnakeConfig{
				Duration:        time.Second * 30,
				Palette:         golds,
				Speed:           20,
				RandomizeColors: true,
				Head:            1,
				NumSnakes:       45,
				SnakeLength:     80,
			}),
		},
		// {
		// 	Label:    "test flood fill",
		// 	Offset:   time.Second,
		// 	Duration: time.Second * 2,
		// 	Effect: effects.NewFloodFill(sys.DebugGetLEDByCoord(0.5, 0.0, 0.5),
		// 		100, colorful.Color{0, 1, 0}, effects.FadeOutFade, 0.5, 0.9, 50,
		// 		ease.OutExpo),
		// },
		// {
		// 	Label:    "testing falling beads",
		// 	Offset:   0,
		// 	Duration: time.Second * 30,
		// 	Effect:   effects.NewFallingBeads(),
		// },
		// {
		// 	Label:    "test flood fill",
		// 	Offset:   time.Second,
		// 	Duration: time.Second * 2,
		// 	Effect: effects.NewFloodFill(sys.DebugGetLEDByCoord(0.5, 0.0, 0.5),
		// 		100, colorful.Color{0, 1, 0}, effects.FadeOutFade, 0.5, 0.9, 50,
		// 		ease.OutExpo),
		// },
		// {
		// 	Label:    "test flood fill",
		// 	Offset:   time.Second,
		// 	Duration: time.Second * 2,
		// 	Effect: effects.NewFloodFill(sys.DebugGetLEDByCoord(0.5, 0.0, 0.5),
		// 		100, colorful.Color{0, 1, 0}, effects.FadeOutRipple, 0.5, 0.9, 50),
		// },
		{
			Label:    "good snake settings",
			Offset:   0,
			Duration: time.Minute * 5,
			Effect: NewAvoidingSnake(&AvoidingSnakeConfig{
				Duration:        time.Minute * 5,
				Palette:         golds,
				Speed:           20,
				RandomizeColors: true,
				Head:            1,
				NumSnakes:       45,
				SnakeLength:     80,
			}),
		},
		// {
		// 	Label:    "good snake settings",
		// 	Offset:   0,
		// 	Duration: time.Minute * 5,
		// 	Effect: effects.NewAvoidingSnake(&effects.AvoidingSnakeConfig{
		// 		Duration:        time.Minute * 5,
		// 		Palette:         golds,
		// 		Speed:           20,
		// 		RandomizeColors: true,
		// 		Head:            1,
		// 		NumSnakes:       45,
		// 		SnakeLength:     80,
		// 	}),
		// },
		// {
		// 	Label:    "display white for 10 seconds as background layer",
		// 	Offset:   0,                // start at 0 seconds
		// 	Duration: time.Second * 10, // end at 10 seconds
		// 	Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
		// 		led.Color = colorful.Color{led.X * (1 - p), led.Y * (1 - p), led.Z * (1 - p)} // just make all LEDs white
		// 	}),
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    1,
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    2,
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    3,
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    1,
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    2,
		// },
		// {
		// 	Label:    "snake effect",
		// 	Offset:   0,
		// 	Duration: time.Hour,
		// 	Effect:   effects.NewSnake(time.Hour, 50, golds[rand.Intn(len(golds))]),
		// 	Layer:    3,
		// },
		// {
		// 	Label:    "display white for 10 seconds as background layer",
		// 	Offset:   0,                   // start at 0 seconds
		// 	Duration: time.Second * 10000, // end at 10 seconds
		// 	Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
		// 		led.Color = colorful.Color{led.X, led.Y, led.Z} // just make all LEDs white
		// 	}),
		// },
		// {
		// 	Label:    "flash blue 10 times as foreground layer",
		// 	Offset:   0,                // start at 0 seconds
		// 	Duration: time.Second * 10, // end at 10 seconds
		// 	Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) { // create a blendable effect (function)
		// 		return colorful.Color{0, 0, 1}, p // blue, with p (progress, from [0..1] representing the progress of the animation) as the blending factor
		// 	}), ledsim.BlendLuvLCh). // use LuvLCh blending
		// 					WithEasing(ease.OutCubic). // ease the progress function with OutCubic
		// 					WithRepetition(20, true),  // repeat 10 times, with reversing (so it animates the flashing on and flashing off)
		// 	Layer: 1, // render this after the white (which is layer 0)
		// },
		// {
		// 	Label:    "red",
		// 	Offset:   time.Second * 10,
		// 	Duration: time.Minute,
		// 	Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) {
		// 		return colorful.Color{1, 0, 0}, 1
		// 	}), ledsim.BlendLuvLCh),
		// },
		// {
		// 	Label:    "green",
		// 	Offset:   time.Second * 10,
		// 	Duration: time.Minute,
		// 	Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) {
		// 		return colorful.Color{0, 1, 0}, 0.5
		// 	}), ledsim.BlendLuvLCh),
		// },
	}
}
