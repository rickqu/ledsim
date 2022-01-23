package parameters

import (
	"github.com/lucasb-eyer/go-colorful"
)

var Params = []Parameter{
	&SlideParam{
		Name:            "Brightness",
		Value:           50,
		LowerBoundLabel: "Dimmer",
		UpperBoundLabel: "Brighter",
	},
	&ColourParam{
		Name: "Gold",
		Color: colorful.Color{
			R: 230,
			G: 190,
			B: 138,
		},
	},
	&ThemeParam{
		Name:           "Season",
		Value:          "Spring",
		PossibleValues: []string{"Summer", "Autumn", "Winter", "Spring"},
	},
}

var ArtworkInformation = ArtworkInfo{
	ArtworkName: "Kintsugi",
}
