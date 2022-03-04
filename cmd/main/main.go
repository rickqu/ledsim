package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"ledsim"
	"ledsim/control_panel"
	"ledsim/effects"
	"ledsim/generator"
	"ledsim/mpv"
	"ledsim/outputs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lucasb-eyer/go-colorful"
)

const frameRate = 10

func main() {
	sys := ledsim.NewSystem()
	ledsim.LoadLEDs(sys)

	var player *mpv.Player
	var err error
	if len(os.Args) >= 2 {
		player, err = mpv.NewPlayer(os.Args[1], os.Getenv("MPV_ARGS"), len(os.Args) >= 3)
		if err != nil {
			panic(err)
		}
		defer player.Close()
	} else {
		log.Println("warn: running without audio/mpv")
	}

	timings, err := generator.ParseTimings(bytes.NewReader(ledsim.TimingData))
	if err != nil {
		panic(fmt.Errorf("parse timings: %w", err))
	}

	gen := generator.NewGenerator([]generator.GeneratableEffect{
		// effects.AvoidingSnakeGenerator,
		effects.SparkleGenerator,
		effects.SegmentGenerator,
		// effects.PulseGenerator,
		// effects.FillUpGenerator,
	})
	keyframes := gen.Generate(timings, time.Now().UnixNano()) // generate some effects

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
	// metrics.StartMetrics()

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
		{250, 130, 0},
		// {153, 101, 21},
		// {244, 163, 0},
	}

	for i, gold := range golds {
		gold.R = gold.R / 255.0
		gold.G = gold.G / 255.0
		gold.B = gold.B / 255.0
		golds[i] = gold
		// log.Println(gold)
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
	// 	log.Println("profile complete")
	// }()

	var getTimestamp func() (time.Duration, error)
	if player != nil {
		getTimestamp = player.GetTimestamp
	}

	offsets := []time.Duration{0, 30, 60, 90, 120, 150}
	for i, _ := range offsets {
		offsets[i] = offsets[i] * time.Second
	}

	mainEffects := []*ledsim.Keyframe{
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
		//{
		//	Label:    "sparkle test",
		//	Offset:   offsets[0],
		//	Duration: offsets[1] - offsets[0],
		//	Effect:   effects.NewSparkle(offsets[1]-offsets[0], time.Second*3, time.Second*3, golds[0]),
		//},
		//{
		//	Label:    "snake fade in",
		//	Offset:   offsets[1],
		//	Duration: time.Second * 5,
		//	Effect:   effects.NewFadeTransition(effects.FADE_IN),
		//	Layer:    2,
		//},
		//{
		//	Label:    "good snake settings",
		//	Offset:   offsets[1],
		//	Duration: offsets[2] - offsets[1],
		//	Effect: effects.NewAvoidingSnake(&effects.AvoidingSnakeConfig{
		//		Duration:        offsets[2] - offsets[1],
		//		Palette:         golds,
		//		Speed:           20,
		//		RandomizeColors: true,
		//		Head:            1,
		//		NumSnakes:       45,
		//		SnakeLength:     80,
		//	}),
		//	Layer: 0,
		//},
		//{
		//	Label:    "snake fade out",
		//	Offset:   offsets[2] - time.Second*5,
		//	Duration: time.Second * 5,
		//	Effect:   effects.NewFadeTransition(effects.FADE_OUT),
		//	Layer:    2,
		//},
		//{
		//	Label:    "idle fade in",
		//	Offset:   offsets[2],
		//	Duration: time.Second * 5,
		//	Effect:   effects.NewFadeTransition(effects.FADE_IN),
		//	Layer:    2,
		//},
		//{
		//	Label:    "idle",
		//	Offset:   offsets[2],
		//	Duration: offsets[3] - offsets[2],
		//	Effect:   effects.NewMonocolour(golds[1]),
		//	Layer:    0,
		//},
		//{
		//	Label:    "idle sparkle",
		//	Offset:   offsets[2],
		//	Duration: offsets[3] - offsets[2],
		//	Effect:   effects.NewSparkle(20*time.Second, time.Second*3, time.Second*3, colorful.Color{255, 255, 255}),
		//	Layer:    1,
		//},
		//{
		//	Label:    "idle fade out",
		//	Offset:   offsets[3] - time.Second*5,
		//	Duration: time.Second * 5,
		//	Effect:   effects.NewFadeTransition(effects.FADE_OUT),
		//	Layer:    2,
		//},
		//{
		//	Label:    "segment in",
		//	Offset:   offsets[3],
		//	Duration: offsets[4] - offsets[3],
		//	Effect:   effects.NewSegment(&golds[0], effects.FADE_IN),
		//},
		//{
		//	Label:    "segment out",
		//	Offset:   offsets[4],
		//	Duration: offsets[5] - offsets[4],
		//	Effect:   effects.NewSegment(&golds[0], effects.FADE_OUT),
		//},
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
	}

	_ = mainEffects

	testEffects := []*ledsim.Keyframe{
		// {
		// 	Label:    "background",
		// 	Offset:   0,
		// 	Duration: time.Second * 10,
		// 	Effect:   effects.NewMonocolour(colorful.Color{0, 0, 0}),
		// },
		// {
		// 	Label:    "sparkle test",
		// 	Offset:   0,
		// 	Duration: time.Second * 5,
		// 	Effect:   effects.NewSparkle(time.Second*5, time.Second*3, time.Second*3, golds[0]),
		// },
		// {
		// 	Label:    "fill up test",
		// 	Offset:   time.Second * 5,
		// 	Duration: time.Second * 5,
		// 	Effect:   effects.NewFillUp(time.Second*4, time.Second, 0.1, golds[0]),
		// },
	}

	_ = mainEffects
	_ = testEffects

	pipeline := []ledsim.Middleware{
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(keyframes), getTimestamp),
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
	pipeline = append(pipeline, ledsim.NewOutput(outputs.NewTeensyNetwork(e, sys)))

	executor := ledsim.NewExecutor(sys, frameRate, pipeline...) // ledsim.TimingStats{},
	// ledsim.StallCheck{},

	timeout := 650 * time.Second
	if player == nil {
		timeout = 643 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if ctx.Err() != nil {
				log.Println("emergency shutdown")
				os.Exit(1)
			}

			log.Println("ctrl+c received, quitting...")
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

	log.Println("running")

	if player != nil {
		go func() {
			t := time.NewTicker(time.Millisecond * 500)
			for {
				select {
				case <-t.C:
					dur, err := player.GetTimestamp()
					if err != nil {
						continue
					}

					if dur >= 642*time.Second {
						log.Println("reached end of file, quitting...")
						cancel()
						return
					}
				case <-ctx.Done():
					t.Stop()
					return
				}
			}
		}()
	}

	err = executor.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Println("executor error:", err)
	}
	log.Println("execution ended")

	// TODO: play grace period animation?

	//log.Println("playing grace period animation")
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//pipeline[0] = ledsim.NewEffectsRunner(ledsim.NewEffectsManager([]*ledsim.Keyframe{}), nil),
	//
	//	executor := ledsim.NewExecutor(sys, frameRate, pipeline...)
	//err = executor.Run()

}
