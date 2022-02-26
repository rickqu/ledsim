package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"ledsim"
	"ledsim/control_panel"
	"ledsim/control_panel/parameters"
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
		{200, 250, 0},
		{200 / 10, 250 / 10, 0},
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

	offsets := []time.Duration{0, 30, 60, 90, 120, 150}
	for i, _ := range offsets {
		offsets[i] = offsets[i] * time.Second
	}

	varColours := []*colorful.Color{
		&parameters.GetParameter("Colour").(*parameters.ColourParam).Color,
		&parameters.GetParameter("Colour2").(*parameters.ColourParam).Color,
	}

	pipeline := []ledsim.Middleware{
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			[]*ledsim.Keyframe{
				{
					Label:    "idle",
					Offset:   offsets[0],
					Duration: offsets[1] - offsets[0],
					Effect:   effects.NewMonocolour(varColours[0]),
					Layer:    0,
				},
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
	pipeline = append(pipeline, ledsim.NewOutput(outputs.NewTeensyNetwork(e, sys)))

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
