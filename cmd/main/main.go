package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"ledsim"
	"ledsim/control_panel"
	"ledsim/effects"
	"ledsim/metrics"
	"ledsim/mpv"
	"ledsim/outputs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	metrics.StartMetrics()
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

	var getTimestamp func() (time.Duration, error)
	if player != nil {
		getTimestamp = player.GetTimestamp
	}

	pipeline := []ledsim.Middleware{
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			effects.GetEffects(),
		), getTimestamp),
		ledsim.NewOutput(mirage),
	}

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
