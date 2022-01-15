package parameters

import (
	"errors"
	"reflect"

	"github.com/lucasb-eyer/go-colorful"
)

type ArtworkInfo struct {
	ArtworkName string
}

type Parameter interface {
	GetName() string
}

type SlideParam struct {
	Name            string
	Value           int
	LowerBoundLabel string
	UpperBoundLabel string
}

func (s SlideParam) GetName() string {
	return s.Name
}

type ColourParam struct {
	Name string
	colorful.Color
}

func (s ColourParam) GetName() string {
	return s.Name
}

type ThemeParam struct {
	Name           string
	Value          string
	PossibleValues []string
}

func (s ThemeParam) GetName() string {
	return s.Name
}

var params []Parameter

func LoadParams() {
	params = Params
}

func GetParameters() []Parameter {
	return params
}

func GetArtworkInfo() ArtworkInfo {
	return ArtworkInformation
}

func SetParam(command *SetParamCommand) error {
	for i := range params {
		if params[i].GetName() == command.ParamName {
			err := parseParamUpdate(i, command.Param)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("No param found")
}

func parseParamUpdate(index int, paramToParse interface{}) error {
	switch params[index].(type) {
	case SlideParam:
		newParamValue, ok := paramToParse.(float64) // in JSON there is only Number type, hence it converts to float64 in Go
		if !ok {
			return errors.New("Tried to parse SlideParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := params[index].(SlideParam)
		updatedParam.Value = int(newParamValue)
		params[index] = updatedParam
		return nil
	case ColourParam:
		newParamValue, ok := paramToParse.(map[string]interface{})
		if !ok {
			return errors.New("Tried to parse ColourParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := params[index].(ColourParam)
		err := updateColourFromInput(&updatedParam, newParamValue)
		if err != nil {
			return err
		}
		params[index] = updatedParam
		return nil
	case ThemeParam:
		newParamValue, ok := paramToParse.(string)
		if !ok {
			return errors.New("Tried to parse ThemeParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := params[index].(ThemeParam)
		updatedParam.Value = newParamValue
		params[index] = updatedParam
		return nil
	default:
		return errors.New("Could not parse type " + reflect.TypeOf(params[index]).Name())
	}
}

func updateColourFromInput(colourParam *ColourParam, inputColour map[string]interface{}) error {
	var ok bool
	colourParam.R, ok = inputColour["R"].(float64)
	if !ok {
		return errors.New("Failed to parse value for colour component R")
	}
	colourParam.G, ok = inputColour["G"].(float64)
	if !ok {
		return errors.New("Failed to parse value for colour component G")
	}
	colourParam.B, ok = inputColour["B"].(float64)
	if !ok {
		return errors.New("Failed to parse value for colour component B")
	}
	return nil
}
