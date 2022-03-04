//go:build ignore
// +build ignore

package effects

import (
	"ledsim"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func ColorFadeGenerator(fadeIn, effect, fadeOut time.Duration, rng *rand.Rand) []*ledsim.Keyframe {
	totalTime := fadeIn + effect + fadeOut
	repeats := math.Round(float64(totalTime) / float64(StandardPeriod/2))
	playTime := time.Duration(float64(totalTime) / repeats)

	var keyframes []*ledsim.Keyframe

	keyframes = append(keyframes,
		&ledsim.Keyframe{
			Label:    "ColorFade_FadeIn_" + uuid.New().String(),
			Offset:   0,
			Duration: fadeIn,
			Effect:   NewFadeTransition(FADE_IN),
			Layer:    2,
		},
		&ledsim.Keyframe{
			Label:    "ColorFade_FadeOut_" + uuid.New().String(),
			Offset:   fadeIn + effect,
			Duration: fadeOut,
			Effect:   NewFadeTransition(FADE_OUT),
			Layer:    2,
		},
	)

	for i := 0; i < int(repeats); i++ {
		col := Golds[i%len(Golds)]

		keyframes = append(keyframes,
			&ledsim.Keyframe{
				Label:    "ColorFade_Main_" + strconv.Itoa(i) + "_" + uuid.New().String(),
				Offset:   time.Duration(i) * playTime,
				Duration: playTime,
				Effect:   NewMonocolour(),
				Layer:    1,
			},
		)
	}

	return keyframes
}
