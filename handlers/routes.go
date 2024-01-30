package handlers

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"net/http"
	"sort"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	return e.String(http.StatusOK, "OK")
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
				Reading:        current.Notes,
				Certificate:    current.Certificate,
			},
		}
	}

	return e.JSON(200, currPayload)
}

func (hl handlers) PracticeAnnounceAPI(e echo.Context) error {
	startParam := e.QueryParam("start")
	amountParam := e.QueryParam("amount")
	facultyParam := e.QueryParam("faculty")

	start, err := strconv.Atoi(startParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid start parameter")
	}

	amount, err := strconv.Atoi(amountParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid amount parameter")
	}

	var sortedStudents []interface{}
	sortedStudents = append(sortedStudents, nil)
	for _, student := range hl.Controller.StudentList {
		if student.Faculty != facultyParam {
			continue
		}
		sortedStudents = append(sortedStudents, student)
	}

	sort.SliceStable(sortedStudents, func(i, j int) bool {
		if sortedStudents[i] == nil || sortedStudents[j] == nil {
			return false
		}
		studentI := sortedStudents[i].(entity.Student)
		studentJ := sortedStudents[j].(entity.Student)
		return studentI.OrderOfReceive < studentJ.OrderOfReceive
	})
	sortedStudents = append(sortedStudents, nil)

	var previousStudent *entity.Student
	var payloads []entity.IndividualPayload

	for i, student := range sortedStudents {

		if i >= start && i < start+amount {
			if student == nil {
				payloads = append(payloads, entity.IndividualPayload{})
				continue
			}
			student := student.(entity.Student)
			var certificateValue string
			if previousStudent != nil {
				if student.Major != previousStudent.Major ||
					student.Degree != previousStudent.Degree ||
					student.Honor != previousStudent.Honor {
					certificateValue = student.Certificate
				} else {
					certificateValue = ""
				}
			} else {
				certificateValue = student.Certificate // First student case
			}

			payload := entity.IndividualPayload{
				Type: "student name",
				Data: entity.StudentPayload{
					OrderOfReading: student.OrderOfReceive,
					Name:           student.FirstName + " " + student.LastName,
					Reading:        student.Notes,
					Certificate:    certificateValue,
				},
			}
			payloads = append(payloads, payload)
			prev := student
			previousStudent = &prev
		}
	}

	return e.JSON(http.StatusOK, payloads)
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/", hl.Mainpage)
	e.GET("/api/announce", hl.AnnounceAPI)
	e.GET("/api/practice/announce", hl.PracticeAnnounceAPI)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

}
