package control_panel

import (
	"encoding/json"
	"ledsim/control_panel/parameters"
	"net/http"

	"github.com/labstack/echo/v4"
)

func InitControlPanel(e *echo.Echo) {
	e.GET("/control/params", func(c echo.Context) error {
		params := parameters.GetParameters()
		paramsJson, err := json.Marshal(params)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, string(paramsJson))
	})
}
