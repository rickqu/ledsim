package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"ledsim"
	"ledsim/control_panel"
	"ledsim/effects"
	"ledsim/mpv"
	"ledsim/outputs"
	"log"
	"net/http"
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

	var player *mpv.Player
	var err error
	if len(os.Args) >= 2 {
		player, err = mpv.NewPlayer(os.Args[1], len(os.Args) >= 3)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("warn: running without audio/mpv")
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/seek/:duration", func(c echo.Context) error {
		if player == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "seeking not supported when no audio is playing")
		}

		duration, err := time.ParseDuration(c.Param("duration"))
		if err != nil {
			return err
		}

		err = player.SeekTo(duration)
		if err != nil {
			return err
		}

		return c.String(http.StatusOK, "seek success")
	})

	control_panel.InitControlPanel(e)

	mirage := outputs.NewMirage(e)

	go func() {
		log.Fatalln(e.Start(":9000"))
	}()

	// {110, 250, 0} comes out as green, however our LEDs render it as
	// gold. The colour we actually want is  {230, 190, 138}
	golds := []colorful.Color{
		// {255, 255, 0},
		// {212, 175, 55},
		// {207, 181, 59},
		// {197, 179, 88},
		{110, 250, 0},
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

	// f, ferr := os.Create("./cpu.prof")
	// if ferr != nil {
	// 	log.Fatal(ferr)
	// }
	// pprof.StartCPUProfile(f)

	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	pprof.StopCPUProfile()
	// 	f.Close()
	// 	fmt.Println("profile complete")
	// }()

	var getTimestamp func() (time.Duration, error)
	if player != nil {
		getTimestamp = player.GetTimestamp
	}

	pipeline := []ledsim.Middleware{
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			[]*ledsim.Keyframe{
				// {
				// 	Label:    "shooting star test",
				// 	Offset:   0,
				// 	Duration: time.Second,
				// 	Effect:   effects.NewShootingStar(effects.Vector{0, 0, 0}, effects.Vector{1, 1, 1}),
				// },
				// {
				// 	Label:    "segment test",
				// 	Offset:   0,
				// 	Duration: time.Second * 30,
				// 	Effect:   effects.NewSegmentShift(time.Second*5, 50, 30, 70, golds[0]),
				// },
				//
				// BEGIN My monocolour test
				//
				// {
				// 	Label:    "monocolour fade in",
				// 	Offset:   0,
				// 	Duration: time.Second * 3,
				// 	Effect:   effects.NewFadeTransition(effects.FADE_IN),
				// 	Layer:    2,
				// },
				// {
				// 	Label:    "monocolour",
				// 	Offset:   0,
				// 	Duration: time.Second * 10,
				// 	Effect:   effects.NewMonocolour(colorful.Color{R: 0.5, G: 0.6, B: 0.7}),
				// },
				// {
				// 	Label:    "monocolour fade out",
				// 	Offset:   7 * time.Second,
				// 	Duration: time.Second * 3,
				// 	Effect:   effects.NewFadeTransition(effects.FADE_OUT),
				// 	Layer:    2,
				// },
				//
				// END My monocolour test
				//
				{
					Label:    "sparkle test",
					Offset:   0,
					Duration: time.Second * 20,
					// duration, baseline, deviation time.Duration, target colorful.Color
					Effect: effects.NewSparkle(20*time.Second, time.Second*3, time.Second*3, golds[0]),
				},
				{
					Label:    "snake fade in",
					Offset:   time.Second * 20,
					Duration: time.Second * 5,
					Effect:   effects.NewFadeTransition(effects.FADE_IN),
					Layer:    2,
				},
				{
					Label:    "good snake settings",
					Offset:   time.Second * 20,
					Duration: time.Second * 30,
					Effect: effects.NewAvoidingSnake(&effects.AvoidingSnakeConfig{
						Duration:        time.Second * 30,
						Palette:         golds,
						Speed:           20,
						RandomizeColors: true,
						Head:            1,
						NumSnakes:       45,
						SnakeLength:     80,
					}),
					Layer: 0,
				},
				{
					Label:    "snake fade out",
					Offset:   45 * time.Second,
					Duration: time.Second * 5,
					Effect:   effects.NewFadeTransition(effects.FADE_OUT),
					Layer:    2,
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
			},
		), getTimestamp),
		ledsim.NewOutput(mirage),
	}

	// genRange := func(start, end int) []int {
	// 	result := make([]int, end-start)
	// 	for i := start; i < end; i++ {
	// 		result[i-start] = i
	// 	}
	// 	return result
	// }

	// udpOutput, err := outputs.NewUDP("192.168.0.1:5050", map[string][]int{
	// 	"192.168.0.2:8888": genRange(0, 300),
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// pipeline = append(pipeline, ledsim.NewOutput(udpOutput))
	// pipeline = append(pipeline, ledsim.NewOutput(outputs.NewTeensyNetwork(e, sys)))

	executor := ledsim.NewExecutor(sys, 30, pipeline...) // ledsim.TimingStats{},
	// ledsim.StallCheck{},

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
			if player != nil {
				player.Close()
			}
			cancel()
		}
	}()

	if player != nil {
		err = player.Play()
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("running")

	err = executor.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Println("executor error:", err)
	}
	fmt.Println("execution ended")
}
