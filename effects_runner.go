package ledsim

import (
	"sort"
	"time"
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
	keyframeBuckets [][]*Keyframe // each bucket is 1 second
	lastKeyframes   []*Keyframe
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
		// bucketN := len(keyframeBuckets) - 1
		// fmt.Println("keyframe bucket:", bucketN)
		// for _, keyframe := range bucket {
		// 	fmt.Println(keyframe.Label)
		// }
		// fmt.Println("")

	}

	return &EffectsManager{
		keyframeBuckets: keyframeBuckets,
		lastKeyframes:   []*Keyframe{},
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

func (r *EffectsManager) Evaluate(system *System, start time.Time, now time.Time) {
	delta := now.Sub(start)

	bucketNum := int(delta / bucketSize)
	if bucketNum >= len(r.keyframeBuckets) {
		return
	}

	bucket := r.keyframeBuckets[bucketNum]

	currentKeyframes := make([]*Keyframe, 0, len(bucket))

	for _, keyframe := range bucket {
		if delta >= keyframe.Offset && delta < keyframe.EndOffset() {
			currentKeyframes = append(currentKeyframes, keyframe)
		}
	}

	for _, lastKeyframe := range r.lastKeyframes {
		if !isKeyframeIn(lastKeyframe, currentKeyframes) {
			// exiting keyframe
			lastKeyframe.Effect.OnExit(system)
		}
	}

	for _, keyframe := range currentKeyframes {
		if !isKeyframeIn(keyframe, r.lastKeyframes) {
			// entering keyframe
			keyframe.Effect.OnEnter(system)
		}
	}

	r.lastKeyframes = currentKeyframes

	for _, keyframe := range currentKeyframes {
		progress := float64(delta-keyframe.Offset) / float64(keyframe.Duration)

		keyframe.Effect.Eval(progress, system)
	}
}

type EffectsRunner struct {
	manager *EffectsManager
	start   time.Time
}

func NewEffectsRunner(manager *EffectsManager) *EffectsRunner {
	return &EffectsRunner{
		manager: manager,
		start:   time.Now(),
	}
}

func (e *EffectsRunner) Execute(system *System, next func() error) error {
	e.manager.Evaluate(system, e.start, time.Now())
	return next()
}
