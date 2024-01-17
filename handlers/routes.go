package handlers

import (
	conx "capstone/server/controller"
	"capstone/server/entity"

	"github.com/labstack/echo/v4"
)

type handlers struct {
	Controller conx.Controller
	Echo       *echo.Echo
}

func NewHandlers(controller conx.Controller) handlers {
	return handlers{
		Controller: controller,
		Echo:       echo.New(),
	}
}

func (hl handlers) Healthcheck(e echo.Context) error {
	return e.String(200, "OK")
}

func (hl handlers) Mainpage(e echo.Context) error {
	return e.File("html/index.html")
}

func (hl handlers) AnnounceAPI(e echo.Context) error {
	var currPayload entity.IndividualPayload
	if currentStudent, ok := hl.Controller.StudentList[hl.Controller.GlobalCounter]; ok {
		current := currentStudent
		currPayload = entity.IndividualPayload{
			Type: "student name",
			Data: entity.StudentPayload{
				OrderOfReading: hl.Controller.GlobalCounter,
				Name:           current.FirstName + " " + current.LastName,
				Reading:        current.Certificate,
				Note:           current.Notes,
				Certificate:    current.Certificate,
			},
		}
	}

	return e.JSON(200, currPayload)
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/", hl.Mainpage)
	e.GET("/api/announce", hl.AnnounceAPI)
}
