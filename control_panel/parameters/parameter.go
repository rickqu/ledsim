package parameters

import (
	"github.com/lucasb-eyer/go-colorful"
)

type Parameter interface{}

type SlideParam struct {
	Name            string
	Value           int32
	LowerBoundLabel string
	UpperBoundLabel string
}

type ColourParam struct {
	Name  string
	Value colorful.Color
}

type ThemeParam struct {
	Name           string
	Value          string
	PossibleValues []string
}
