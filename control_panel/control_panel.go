package control_panel

import (
	"encoding/json"
	"ledsim/control_panel/parameters"
	"net/http"

	"github.com/labstack/echo/v4"
)

const CONTROL_SUBDIRECTORY = "/control"

func InitControlPanel(e *echo.Echo) {
	parameters.LoadParams()
	e.GET(CONTROL_SUBDIRECTORY+"/artwork_info", func(c echo.Context) error {
		artworkInfo, err := json.Marshal(parameters.GetArtworkInfo())
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, string(artworkInfo))
	})

	e.GET(CONTROL_SUBDIRECTORY+"/params", func(c echo.Context) error {
		params := parameters.GetParameters()
		paramsJson, err := json.Marshal(params)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, string(paramsJson))
	})

	e.POST(CONTROL_SUBDIRECTORY+"/set_param", func(c echo.Context) error {
		setParamCommand := new(parameters.SetParamCommand)
		if err := c.Bind(setParamCommand); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		if err := parameters.SetParam(setParamCommand); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		return c.String(http.StatusOK, "")
	})
}
