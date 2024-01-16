package handlers

import (
	"capstone/server/entity"

	"github.com/labstack/echo/v4"
)

type Controller struct {
	GlobalCounter        int
	MicrocontrollerAlive bool
	StudentList          map[int]entity.Student
}

func (c Controller) Healthcheck(e echo.Context) error {
	return e.String(200, "OK")
}

func (c Controller) Mainpage(e echo.Context) error {
	return e.File("html/index.html")
}

func (c Controller) AnnounceAPI(e echo.Context) error {
	var currPayload entity.IndividualPayload
	if currentStudent, ok := c.StudentList[c.GlobalCounter]; ok {
		current := currentStudent
		currPayload = entity.IndividualPayload{
			Type: "student name",
			Data: entity.StudentPayload{
				OrderOfReading: c.GlobalCounter,
				Name:           current.FirstName + " " + current.LastName,
				Reading:        current.Certificate,
				Note:           current.Notes,
				Certificate:    current.Certificate,
			},
		}
	}

	return e.JSON(200, currPayload)
}

func (c Controller) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", c.Healthcheck)
	e.GET("/", c.Mainpage)
	e.GET("/api/announce", c.AnnounceAPI)
}
