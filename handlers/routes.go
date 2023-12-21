package handlers

import (
	"github.com/labstack/echo/v4"
)

func Healthcheck(e echo.Context) error {
	return e.String(200, "OK")
}

func Mainpage(e echo.Context) error {
	return e.File("html/index.html")
}

func RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", Healthcheck)
	e.GET("/", Mainpage)
}
