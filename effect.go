package ledsim

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

type Effect interface {
	OnEnter(system *System)
	Eval(progress float64, system *System)
	OnExit(system *System)
}

type BlendableEffect interface {
	OnEnter(system *System)
	BlendEval(progress float64, led *LED) (colorful.Color, float64)
	OnExit(system *System)
}

type EffectFunc func(progress float64, system *System)

func (f EffectFunc) OnEnter(system *System) {
}

func (f EffectFunc) OnExit(system *System) {
}

func (f EffectFunc) Eval(progress float64, system *System) {
	f(progress, system)
}

type LEDEffect func(progress float64, led *LED)

func (f LEDEffect) OnEnter(system *System) {
}

func (f LEDEffect) OnExit(system *System) {
}

func (f LEDEffect) Eval(progress float64, system *System) {
	for _, led := range system.LEDs {
		f(progress, led)
	}
}

type BlendableEffectFunc func(progress float64, led *LED) (colorful.Color, float64)

func (f BlendableEffectFunc) OnEnter(system *System) {
}

func (f BlendableEffectFunc) OnExit(system *System) {
}

func (f BlendableEffectFunc) BlendEval(progress float64, led *LED) (colorful.Color, float64) {
	return f(progress, led)
}

type Blending func(from colorful.Color, to colorful.Color, t float64) colorful.Color

type blendingWrapper struct {
	effect   BlendableEffect
	blending Blending
}

func (w *blendingWrapper) OnEnter(system *System) {
	w.effect.OnEnter(system)
}

func (w *blendingWrapper) OnExit(system *System) {
	w.effect.OnExit(system)
}

func (w *blendingWrapper) Eval(progress float64, system *System) {
	for _, led := range system.LEDs {
		c, blend := w.effect.BlendEval(progress, led)
		led.Color = w.blending(led.Color, c, blend)
	}
}

func NewBlendingEffect(effect BlendableEffect, blending Blending) WrappedEffect {
	return WrappedEffect{&blendingWrapper{
		effect:   effect,
		blending: blending,
	}}
}

type easingWrapper struct {
	effect Effect
	easing func(progress float64) float64
}

func (w *easingWrapper) OnEnter(system *System) {
	w.effect.OnEnter(system)
}

func (w *easingWrapper) OnExit(system *System) {
	w.effect.OnExit(system)
}

func (w *easingWrapper) Eval(progress float64, system *System) {
	w.effect.Eval(w.easing(progress), system)
}

type WrappedEffect struct {
	Effect
}

func (e WrappedEffect) WithEasing(easing func(progress float64) float64) WrappedEffect {
	return WrappedEffect{&easingWrapper{
		effect: e,
		easing: easing,
	}}
}

type repetitionWrapper struct {
	effect Effect
	count  int
}

func (w *repetitionWrapper) OnEnter(system *System) {
	w.effect.OnEnter(system)
}

func (w *repetitionWrapper) OnExit(system *System) {
	w.effect.OnExit(system)
}

func (w *repetitionWrapper) Eval(progress float64, system *System) {
	w.effect.Eval(math.Mod(progress*float64(w.count), 1.0), system)
}

func (e WrappedEffect) WithRepetition(count int, reverse ...bool) WrappedEffect {
	if len(reverse) > 0 && reverse[0] {
		return WrappedEffect{&repetitionWrapper{
			effect: Sequential(e, e.Reverse()),
			count:  count * 2,
		}}
	}

	return WrappedEffect{&repetitionWrapper{
		effect: e,
		count:  count,
	}}
}

type reverseWrapper struct {
	effect Effect
	count  int
}

func (w *reverseWrapper) OnEnter(system *System) {
	w.effect.OnEnter(system)
}

func (w *reverseWrapper) OnExit(system *System) {
	w.effect.OnExit(system)
}

func (w *reverseWrapper) Eval(progress float64, system *System) {
	w.effect.Eval(1.0-progress, system)
}

func (e WrappedEffect) Reverse() WrappedEffect {
	return WrappedEffect{&reverseWrapper{
		effect: e,
	}}
}

type sequentialWrapper struct {
	effects []Effect
}

func (w *sequentialWrapper) OnEnter(system *System) {
	for _, effect := range w.effects {
		effect.OnEnter(system)
	}
}

func (w *sequentialWrapper) OnExit(system *System) {
	for _, effect := range w.effects {
		effect.OnExit(system)
	}
}

func (w *sequentialWrapper) Eval(progress float64, system *System) {
	i := int(math.Floor(progress * float64(len(w.effects))))

	if i >= len(w.effects) {
		i = len(w.effects) - 1
	}

	w.effects[i].Eval(math.Mod(progress*float64(len(w.effects)), 1.0), system)
}

func Sequential(effects ...Effect) WrappedEffect {
	return WrappedEffect{&sequentialWrapper{
		effects: effects,
	}}
}
