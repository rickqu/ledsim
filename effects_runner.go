package ledsim

import (
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"time"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	bucketSize = time.Second
)

type Keyframe struct {
	Label    string
	Offset   time.Duration
	Duration time.Duration
	Effect   Effect
	Layer    int
}

func (k *Keyframe) EndOffset() time.Duration {
	return k.Offset + k.Duration
}

type EffectsManager struct {
	keyframeBuckets  [][]*Keyframe // each bucket is 1 second
	lastKeyframes    []*Keyframe
	blacklist        map[*Keyframe]bool
	lastLoopEnd      time.Duration
	lastDelta        time.Duration
	justFinishedLoop bool
}

var blackFrame = &Keyframe{
	Effect: LEDEffect(func(p float64, led *LED) {
		led.Color = colorful.Color{0, 0, 0}
	}),
}

func NewEffectsManager(keyframes []*Keyframe) *EffectsManager {
	var keyframeBuckets [][]*Keyframe

	for {
		var bucket []*Keyframe
		lower := time.Duration(len(keyframeBuckets)) * bucketSize
		upper := time.Duration(len(keyframeBuckets)+1) * bucketSize

		outOfBounds := true
		// bucketizing algorithm runs in O(n^2), could optimize to run faster.
		for _, keyframe := range keyframes {
			if (keyframe.Offset < lower && keyframe.EndOffset() > lower) ||
				(keyframe.Offset >= lower && keyframe.Offset < upper) {
				bucket = append(bucket, keyframe)
				outOfBounds = false
			} else if keyframe.Offset > upper {
				outOfBounds = false
			}
		}

		if outOfBounds {
			break
		}

		sort.Slice(bucket, func(i, j int) bool {
			return bucket[i].Layer < bucket[j].Layer
		})

		keyframeBuckets = append(keyframeBuckets, bucket)
	}

	return &EffectsManager{
		keyframeBuckets:  keyframeBuckets,
		lastKeyframes:    []*Keyframe{},
		blacklist:        make(map[*Keyframe]bool),
		lastLoopEnd:      0,
		lastDelta:        0,
		justFinishedLoop: false,
	}
}

func isKeyframeIn(needle *Keyframe, haystack []*Keyframe) bool {
	for _, keyframe := range haystack {
		if needle == keyframe {
			return true
		}
	}
	return false
}

var emptyKeyFrames = make([]*Keyframe, 1)

func (r *EffectsManager) Evaluate(system *System, delta time.Duration) {

	var loopTime = delta - r.lastLoopEnd
	bucketNum := int(loopTime / bucketSize)

	if bucketNum == (len(r.keyframeBuckets)) {
		for _, keyFrame := range r.lastKeyframes {
			r.exitAnimations(keyFrame, system)
		}

		// Recalculate the loopTime because we are in a new iteration of animation loop
		r.lastLoopEnd = r.lastDelta
		loopTime = delta - r.lastLoopEnd
		bucketNum = 0
	}
	r.lastDelta = delta

	fmt.Println(loopTime)

	bucket := r.keyframeBuckets[bucketNum]
	currentKeyframes := make([]*Keyframe, 0, len(bucket))

	for _, keyframe := range bucket {
		if loopTime >= keyframe.Offset && loopTime < keyframe.EndOffset() && !r.blacklist[keyframe] {
			currentKeyframes = append(currentKeyframes, keyframe)
		}
	}

	if !r.justFinishedLoop {
		for _, lastKeyframe := range r.lastKeyframes {
			if !isKeyframeIn(lastKeyframe, currentKeyframes) && !r.blacklist[lastKeyframe] {
				r.exitAnimations(lastKeyframe, system)
			}
		}
	}

	for _, keyframe := range currentKeyframes {
		if r.blacklist[keyframe] {
			continue
		}

		if !isKeyframeIn(keyframe, r.lastKeyframes) || r.justFinishedLoop {
			r.enterAnimation(keyframe, system)
		}
	}

	// Clear canvas before running other animations
	blackFrame.Effect.Eval(0, system)

	for _, keyframe := range currentKeyframes {
		if r.blacklist[keyframe] {
			continue
		}

		progress := float64(loopTime-keyframe.Offset) / float64(keyframe.Duration)

		r.runAnimation(keyframe, progress, system)
	}

	r.lastKeyframes = currentKeyframes
	r.justFinishedLoop = false
}

func (r *EffectsManager) enterAnimation(keyframe *Keyframe, system *System) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("warn: panic OnEnter with effect %q: %v\n%s",
				keyframe.Label, rec, string(debug.Stack()))
			log.Printf("warn: %q will be blacklisted", keyframe.Label)
			r.blacklist[keyframe] = true
		}
	}()
	keyframe.Effect.OnEnter(system)
}

func (r *EffectsManager) runAnimation(keyframe *Keyframe, progress float64, system *System) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("warn: panic Eval with effect %q: %v\n%s",
				keyframe.Label, rec, string(debug.Stack()))
			log.Printf("warn: %q will be blacklisted", keyframe.Label)
			r.blacklist[keyframe] = true
		}
	}()

	keyframe.Effect.Eval(progress, system)
}

func (r *EffectsManager) exitAnimations(keyframe *Keyframe, system *System) {
	defer func() {
		if rec := recover(); rec != nil {
			// get stack trace
			log.Printf("warn: panic OnExit with effect %q: %v\n%s",
				keyframe.Label, rec, string(debug.Stack()))
		}
	}()
	keyframe.Effect.OnExit(system)
}

type EffectsRunner struct {
	manager    *EffectsManager
	start      time.Time
	timeGetter func() (time.Duration, error)
}

func NewEffectsRunner(manager *EffectsManager, timeGetter ...func() (time.Duration, error)) *EffectsRunner {
	if len(timeGetter) > 0 {
		return &EffectsRunner{
			manager:    manager,
			timeGetter: timeGetter[0],
			start:      time.Now(),
		}
	}

	return &EffectsRunner{
		manager: manager,
		start:   time.Now(),
	}
}

func (e *EffectsRunner) Execute(system *System, next func() error) error {
	if e.timeGetter != nil {
		t, err := e.timeGetter()
		if err != nil {
			log.Println("warn: error getting time, falling back to wall clock:", err)
			e.manager.Evaluate(system, time.Since(e.start))
		} else {
			e.start = time.Now().Add(-t)
			e.manager.Evaluate(system, t)
		}

		return next()
	}

	e.manager.Evaluate(system, time.Since(e.start))
	return next()
}
