package parameters

import (
	"github.com/lucasb-eyer/go-colorful"
)

var Params = map[string]Parameter{
	"Brightness": &SlideParam{
		Name:            "Brightness",
		Value:           50,
		LowerBoundLabel: "Dimmer",
		UpperBoundLabel: "Brighter",
	},
	"Gold": &ColourParam{
		Name: "Gold",
		Color: colorful.Color{
			R: 230,
			G: 190,
			B: 138,
		},
	},
	"Season": &ThemeParam{
		Name:           "Season",
		Value:          "Spring",
		PossibleValues: []string{"Summer", "Autumn", "Winter", "Spring"},
	},
}

var ArtworkInformation = ArtworkInfo{
	ArtworkName: "Kintsugi",
}
