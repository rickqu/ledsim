// Package generator holds effect generators
package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"ledsim"
)

type GeneratableEffect func(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe

type Generator struct {
	effects []GeneratableEffect
}

func NewGenerator(effects []GeneratableEffect) *Generator {
	return &Generator{
		effects: effects,
	}
}

type Timing struct {
	Offset  time.Duration
	FadeIn  time.Duration
	Effect  time.Duration
	FadeOut time.Duration
}

func (g *Generator) Generate(timings []*Timing, seed int64) []*ledsim.Keyframe {
	rng := rand.New(rand.NewSource(seed))

	var keyframes []*ledsim.Keyframe

	rng.Shuffle(len(g.effects), func(i, j int) {
		g.effects[i], g.effects[j] = g.effects[j], g.effects[i]
	})

	i := 0

	for _, timing := range timings {
		// pick a random effect
		effect := g.effects[i](timing.FadeIn, timing.Effect, timing.FadeOut, rng)

		for _, keyframe := range effect {
			keyframe.Offset += timing.Offset
		}

		keyframes = append(keyframes, effect...)

		i++

		if i >= len(g.effects) {
			i = 0
			lastLast := reflect.ValueOf(g.effects[len(g.effects)-1]).Pointer()

			for {
				rng.Shuffle(len(g.effects), func(i, j int) {
					g.effects[i], g.effects[j] = g.effects[j], g.effects[i]
				})

				first := reflect.ValueOf(g.effects[0]).Pointer()

				if first != lastLast {
					break
				}
			}
		}
	}

	return keyframes
}

var timingPattern = regexp.MustCompile(`^(.+)\t(.+)\t(.+)$`)

func ParseTimings(rd io.Reader) ([]*Timing, error) {
	bufRd := bufio.NewReader(rd)

	type parsedLine struct {
		Start time.Duration
		End   time.Duration
		Type  string
	}

	var parsedLines []*parsedLine

	for {
		line, err := bufRd.ReadBytes('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		matches := timingPattern.FindStringSubmatch(string(line))
		start, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return nil, err
		}

		end, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return nil, err
		}

		parsedLines = append(parsedLines, &parsedLine{
			Start: time.Duration(start * float64(time.Second)),
			End:   time.Duration(end * float64(time.Second)),
			Type:  matches[3],
		})
	}

	var output []*Timing

	for i := 0; i < len(parsedLines); i += 2 {
		fadeIn := parsedLines[i]
		fadeOut := parsedLines[i+1]

		if fadeIn.Type != "fade in" {
			return nil, fmt.Errorf("generator: expected fade in got: %v", fadeIn.Type)
		}

		if fadeOut.Type != "fade out" {
			return nil, fmt.Errorf("generator: expected fade out got: %v", fadeOut.Type)
		}

		output = append(output, &Timing{
			FadeIn:  fadeIn.End - fadeIn.Start,
			Effect:  fadeOut.Start - fadeIn.End,
			FadeOut: fadeOut.End - fadeOut.Start,
			Offset:  fadeIn.Start,
		})
	}

	return output, nil
}
