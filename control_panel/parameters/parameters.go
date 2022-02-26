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
	"Colour": &ColourParam{
		Name: "Colour",
		Color: colorful.Color{
			R: 200,
			G: 250,
			B: 0,
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
