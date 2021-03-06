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

func (s *SlideParam) GetName() string {
	return s.Name
}

type ColourParam struct {
	Name string
	colorful.Color
}

func (s *ColourParam) GetName() string {
	return s.Name
}

type ThemeParam struct {
	Name           string
	Value          string
	PossibleValues []string
}

func (s *ThemeParam) GetName() string {
	return s.Name
}

var params map[string]Parameter

func LoadParams() {
	params = Params
}

func GetParameters() map[string]Parameter {
	return params
}

func GetArtworkInfo() ArtworkInfo {
	return ArtworkInformation
}

func GetParameter(name string) Parameter {
	if paramToReturn, ok := params[name]; ok {
		return paramToReturn
	} else {
		panic("Parameter " + name + " not found")
	}
}

func SetParam(command *SetParamCommand) error {
	for _, value := range params {
		if value.GetName() == command.ParamName {
			err := parseParamUpdate(value, command.Param)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("No param found")
}

func parseParamUpdate(paramValue Parameter, paramToParse interface{}) error {
	switch paramValue.(type) {
	case *SlideParam:
		newParamValue, ok := paramToParse.(float64) // in JSON there is only Number type, hence it converts to float64 in Go
		if !ok {
			return errors.New("Tried to parse SlideParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := paramValue.(*SlideParam)
		updatedParam.GetName()
		updatedParam.Value = int(newParamValue)
		return nil
	case *ColourParam:
		newParamValue, ok := paramToParse.(map[string]interface{})
		if !ok {
			return errors.New("Tried to parse ColourParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := paramValue.(*ColourParam)
		err := updateColourFromInput(updatedParam, newParamValue)
		if err != nil {
			return err
		}
		return nil
	case *ThemeParam:
		newParamValue, ok := paramToParse.(string)
		if !ok {
			return errors.New("Tried to parse ThemeParam but was not successful. Provided value type is " + reflect.TypeOf(paramToParse).Name())
		}
		updatedParam := paramValue.(*ThemeParam)
		updatedParam.Value = newParamValue
		return nil
	default:
		return errors.New("Could not parse type " + reflect.TypeOf(paramValue).Name())
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
	colourParam.R = colourParam.R / 255.0
	colourParam.G = colourParam.G / 255.0
	colourParam.B = colourParam.B / 255.0
	return nil
}
