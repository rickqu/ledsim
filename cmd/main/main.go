package main

import (
	_ "embed"
	"ledsim"
	"ledsim/outputs"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/fogleman/ease"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	scale  = 0.0005 * 2.056422
	origin = [...]float64{
		(1 / (0.0005 * 2.056422)) * -18.04,
		(1 / (0.0005 * 2.056422)) * 9.58,
		(1 / (0.0005 * 2.056422)) * 1,
	}
)

var pattern = regexp.MustCompile(`(?m)^.*{([-\.0-9]+), ([-\.0-9]+), ([-\.0-9]+)}\s*$`)

//go:embed crack_leds.txt
var crackLeds string

func main() {
	sys := ledsim.NewSystem()

	groups := pattern.FindAllStringSubmatch(crackLeds, -1)
	for _, group := range groups {
		x, _ := strconv.ParseFloat(group[1], 64)
		y, _ := strconv.ParseFloat(group[2], 64)
		z, _ := strconv.ParseFloat(group[3], 64)

		sys.AddLED(&ledsim.LED{
			X: -(x - origin[0]) * scale,
			Y: (y - origin[1]) * scale,
			Z: (z - origin[2]) * scale,
		})
	}

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

	executor := ledsim.NewExecutor(sys, 20,
		ledsim.TimingStats{},
		ledsim.NewEffectsRunner(ledsim.NewEffectsManager(
			[]*ledsim.Keyframe{
				{
					Label:    "display white for 10 seconds as background layer",
					Offset:   0,                // start at 0 seconds
					Duration: time.Second * 10, // end at 10 seconds
					Effect: ledsim.LEDEffect(func(p float64, led *ledsim.LED) {
						led.Color = colorful.Color{1, 1, 1} // just make all LEDs white
					}),
				},
				{
					Label:    "flash blue 10 times as foreground layer",
					Offset:   0,                // start at 0 seconds
					Duration: time.Second * 10, // end at 10 seconds
					Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) { // create a blendable effect (function)
						return colorful.Color{0, 0, 1}, p // blue, with p (progress, from [0..1] representing the progress of the animation) as the blending factor
					}), ledsim.BlendLuvLCh). // use LuvLCh blending
									WithEasing(ease.OutCubic). // ease the progress function with OutCubic
									WithRepetition(10, true),  // repeat 10 times, with reversing (so it animates the flashing on and flashing off)
					Layer: 1, // render this after the white (which is layer 0)
				},
				{
					Label:    "red",
					Offset:   time.Second * 10,
					Duration: time.Minute,
					Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) {
						return colorful.Color{1, 0, 0}, 1
					}), ledsim.BlendLuvLCh),
				},
				{
					Label:    "green",
					Offset:   time.Second * 10,
					Duration: time.Minute,
					Effect: ledsim.NewBlendingEffect(ledsim.BlendableEffectFunc(func(p float64, led *ledsim.LED) (colorful.Color, float64) {
						return colorful.Color{0, 1, 0}, 0.5
					}), ledsim.BlendLuvLCh),
				},
			},
		)),
		ledsim.NewOutput(mirage))

	executor.Run()
}
