package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func ok(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, data)
}

func fail(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}
