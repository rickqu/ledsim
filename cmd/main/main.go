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

	executor := ledsim.NewExecutor(sys, 60,
		// ledsim.TimingStats{},
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			[]*ledsim.Keyframe{
				{
					Label:    "display white for 10 seconds as background layer",
					Offset:   0,                   // start at 0 seconds
					Duration: time.Second * 10000, // end at 10 seconds
					Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
						led.Color = colorful.Color{0, 0, 0} // just make all LEDs white
					}),
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    1,
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    2,
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    3,
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    1,
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    2,
				},
				{
					Label:    "snake effect",
					Offset:   0,
					Duration: time.Hour,
					Effect:   effects.NewSnake(time.Hour, 200, colorful.FastHappyColor()),
					Layer:    3,
				},
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
				// 					WithRepetition(2, true),   // repeat 10 times, with reversing (so it animates the flashing on and flashing off)
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
