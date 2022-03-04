package ledsim

import "github.com/lucasb-eyer/go-colorful"

var (
	BlendHcl Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		return from.BlendHcl(to, t)
	}
	BlendHsv Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		return from.BlendHsv(to, t)
	}
	BlendLab Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		return from.BlendLab(to, t)
	}
	BlendLuvLCh Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		return from.BlendLuvLCh(to, t)
	}
	BlendRgb Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		return from.BlendRgb(to, t)
	}
	BlendAdditiveRgb Blending = func(from colorful.Color, to colorful.Color, t float64) colorful.Color {
		from.R += to.R * t
		from.G += to.G * t
		from.B += to.B * t
		return from.Clamped()
	}
)
