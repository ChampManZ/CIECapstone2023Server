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

func (hl handlers) CounterAPI(e echo.Context) error {
	var currPayload entity.CounterPayload
	if _, ok := hl.Controller.StudentList[hl.Controller.GlobalCounter]; ok {
		currPayload = entity.CounterPayload{
			Current:   hl.Controller.GlobalCounter,
			Remaining: len(hl.Controller.StudentList) - hl.Controller.GlobalCounter,
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

func (hl handlers) updateNotes(e echo.Context) error {
	orderOfReceiveParam := e.QueryParam("orderOfReceive")
	noteParam := e.QueryParam("note")

	orderOfReceive, err := strconv.Atoi(orderOfReceiveParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid order of receive parameter")
	}

	err = hl.Controller.MySQLConn.UpdateNote(orderOfReceive, noteParam)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	location := findStudentByOrder(hl.Controller.StudentList, orderOfReceive)
	if location == -1 {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	temp := hl.Controller.StudentList[location]
	temp.Notes = noteParam

	hl.Controller.StudentList[location] = temp
	return e.JSON(http.StatusOK, "OK")
}

func findStudentByOrder(students map[int]entity.Student, order int) int {
	for key, student := range students {
		if student.OrderOfReceive == order {
			return key
		}
	}
	return -1 // Not found
}

func (hl handlers) getFacultiesAPI(e echo.Context) error {
	faculties, err := hl.Controller.MySQLConn.QueryUniqueFaculties()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	return e.JSON(http.StatusOK, faculties)
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/", hl.Mainpage)
	e.GET("/api/announce", hl.AnnounceAPI)
	e.GET("/api/counter", hl.CounterAPI)
	e.GET("/api/practice/announce", hl.PracticeAnnounceAPI)
	e.GET("/api/faculties", hl.getFacultiesAPI)
	e.PUT("/api/notes", hl.updateNotes)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

}
