package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"ledsim"
	"ledsim/effects"
	"ledsim/outputs"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lucasb-eyer/go-colorful"
)

func main() {
	sys := ledsim.NewSystem()
	ledsim.LoadLEDs(sys)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	mirage := outputs.NewMirage(e)

	go func() {
		log.Fatalln(e.Start(":9000"))
	}()

	golds := []colorful.Color{
		{255, 255, 0},
		{212, 175, 55},
		{207, 181, 59},
		{197, 179, 88},
		{230, 190, 138},
		{153, 101, 21},
		{244, 163, 0},
	}

	for i, gold := range golds {
		gold.R = gold.R / 255.0
		gold.G = gold.G / 255.0
		gold.B = gold.B / 255.0
		golds[i] = gold
		// fmt.Println(gold)
	}

	executor := ledsim.NewExecutor(sys, 60,
		// ledsim.TimingStats{},
		ledsim.StallCheck{},
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			[]*ledsim.Keyframe{
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
				// 	Label:    "segment test",
				// 	Offset:   0,
				// 	Duration: time.Minute * 5,
				// 	Effect:   effects.NewSegmentShift(time.Minute*5, 100, 30, 70, golds[0]),
				// },
				{
					Label:    "panfade_pretest",
					Offset:   0,
					Duration: time.Minute * 5,
					Effect:   effects.NewPanningFade(time.Minute*5, 0.5),
				},
				// {
				// 	Label:    "testing avoiding snake",
				// 	Offset:   0,
				// 	Duration: time.Minute * 5,
				// 	Effect: effects.NewAvoidingSnake(&effects.AvoidingSnakeConfig{
				// 		Duration:        time.Minute * 5,
				// 		Palette:         golds,
				// 		Speed:           100,
				// 		RandomizeColors: true,
				// 		Head:            5,
				// 		NumSnakes:       5,
				// 		SnakeLength:     50,
				// 	}),
				// },
				// {
				// 	Label:    "display white for 10 seconds as background layer",
				// 	Offset:   0,                // start at 0 seconds
				// 	Duration: time.Second * 10, // end at 10 seconds
				// 	Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
				// 		led.Color = colorful.Color{led.X, led.Y, led.Z} // just make all LEDs white
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
			},
		)),
		ledsim.NewOutput(mirage))

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if ctx.Err() != nil {
				fmt.Println("emergency shutdown")
				os.Exit(1)
			}

			fmt.Println("ctrl+c received, quitting...")
			cancel()
		}
	}()

	err := executor.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Println("executor error:", err)
	}
	fmt.Println("execution ended")
}
