package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HandleUp(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "up"})
}

func HandleComplete(c echo.Context) error {
	return c.JSON(
		http.StatusOK,
		map[string]string{"msg": "stop bothering me human"},
	)
}
