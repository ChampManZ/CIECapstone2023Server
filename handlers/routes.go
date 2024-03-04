package handlers

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"capstone/server/utility"
	"capstone/server/utility/config"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type handlers struct {
	Controller *conx.Controller
	Echo       *echo.Echo
}

func NewHandlers(controller conx.Controller) handlers {
	return handlers{
		Controller: &controller,
		Echo:       echo.New(),
	}
}

func (hl handlers) Healthcheck(e echo.Context) error {
	return e.String(http.StatusOK, "OK")
}

func (hl handlers) Mainpage(e echo.Context) error {
	return e.File("html/dist/index.html")
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
				Reading:        current.Reading,
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
	var payloads []interface{}

	start, err := strconv.Atoi(startParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid start parameter")
	}

	amount, err := strconv.Atoi(amountParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid amount parameter")
	}

	var found bool
	var filtered_script []entity.IndividualPayload

	//append to front
	filtered_script = append(filtered_script, entity.IndividualPayload{})
	for _, payload := range hl.Controller.Script {
		if payload.Type == "student name" {
			if payload.Data.(entity.StudentPayload).Faculty == facultyParam {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if payload.Type == "script" {
			if payload.Data.(entity.AnnouncerPayload).Faculty == facultyParam {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if found {
			if payload.Type == "student name" {
				if payload.Data.(entity.StudentPayload).Faculty != facultyParam {
					break
				}
			}
			if payload.Type == "script" {
				if payload.Data.(entity.AnnouncerPayload).Faculty != facultyParam {
					break
				}
			}
		}
	}

	//append to back
	filtered_script = append(filtered_script, entity.IndividualPayload{})

	for i, payload := range filtered_script {
		if i >= start && i < start+amount {
			payloads = append(payloads, payload)
		}
	}

	return e.JSON(http.StatusOK, payloads)
}

func (hl handlers) UpdateNotes(e echo.Context) error {
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

	location := utility.FindStudentByOrder(hl.Controller.StudentList, orderOfReceive)
	if location == -1 {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	temp := hl.Controller.StudentList[location]
	temp.Reading = noteParam

	hl.Controller.StudentList[location] = temp
	hl.Controller.GenerateSript()
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) TestscriptAPI(e echo.Context) error {
	hl.Controller.GenerateSript()
	return e.JSON(http.StatusOK, hl.Controller.Script)
}

func (hl handlers) GetFacultiesAPI(c echo.Context) error {
	faculties, err := hl.Controller.MySQLConn.QueryUniqueFaculties()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	groupedFaculties := entity.FacultySessionPayload{}
	for _, faculty := range faculties {
		if faculty.SessionOfAnnounce == "เช้า" {
			groupedFaculties.Morning = append(groupedFaculties.Morning, faculty.Faculty)
		} else if faculty.SessionOfAnnounce == "บ่าย" {
			groupedFaculties.Afternoon = append(groupedFaculties.Afternoon, faculty.Faculty)
		}
	}
	return c.JSON(http.StatusOK, groupedFaculties)
}

func (hl handlers) UpdateStudentList(e echo.Context) error {
	studentData, err := utility.FetchRegistraData(config.GlobalConfig.Download_URL)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	err = hl.Controller.MySQLConn.UpdateStudentList(studentData)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	hl.Controller.StudentList, err = hl.Controller.MySQLConn.QueryStudentsToMap()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	hl.Controller.GenerateSript()
	return e.JSON(http.StatusOK, "OK")

}

func (hl handlers) UpdateAnnouncer(e echo.Context) error {
	var updateRequests []entity.Announcer
	if err := e.Bind(&updateRequests); err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid request format")
	}

	// transaction for safety due to batching update
	tx, err := hl.Controller.MySQLConn.Begin()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	for _, updateRequest := range updateRequests {
		err = hl.Controller.MySQLConn.UpdateAnnouncerQuery(tx, updateRequest.AnnouncerID, updateRequest.AnnouncerName, updateRequest.AnnouncerScript, updateRequest.Session, updateRequest.Start, updateRequest.End)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to commit transaction")
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	hl.Controller.GenerateSript()
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) InsertAnnouncer(e echo.Context) error {
	var insertRequests []entity.Announcer
	if err := e.Bind(&insertRequests); err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid request format")
	}

	// Start a transaction
	tx, err := hl.Controller.MySQLConn.Begin()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	for _, insertRequest := range insertRequests {
		err = hl.Controller.MySQLConn.InsertAnnouncer(tx, insertRequest.AnnouncerName, insertRequest.AnnouncerScript, insertRequest.Session, insertRequest.Start, insertRequest.End)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, fmt.Sprintf("Failed to insert announcer: %v", err))
		}
	}

	if err := tx.Commit(); err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to commit transaction")
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	hl.Controller.GenerateSript()
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) GetAnnouncers(e echo.Context) error {
	announcersMap := hl.Controller.AnnouncerList
	var announcersSlice []entity.Announcer
	for _, announcer := range announcersMap {
		announcersSlice = append(announcersSlice, announcer)
	}
	return e.JSON(http.StatusOK, announcersSlice)
}

func (hl handlers) DeleteAnnouncer(e echo.Context) error {
	var announcerIDs []int
	if err := e.Bind(&announcerIDs); err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid request format")
	}

	// transaction
	tx, err := hl.Controller.MySQLConn.Begin()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to start transaction")
	}
	defer tx.Rollback()

	for _, announcerID := range announcerIDs {
		err = hl.Controller.MySQLConn.DeleteAnnouncer(tx, announcerID)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, fmt.Sprintf("Failed to delete announcer: %v", err))
		}
	}

	if err := tx.Commit(); err != nil {
		return e.JSON(http.StatusInternalServerError, "Failed to commit transaction")
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) DashboardAPI(e echo.Context) error {
	student, ok := hl.Controller.StudentList[hl.Controller.GlobalCounter]
	if !ok {
		payload := entity.DashboardPayload{}
		return e.JSON(http.StatusOK, payload)
	}
	Name := fmt.Sprintf(student.FirstName + " " + student.LastName)
	nextStudent := hl.Controller.StudentList[hl.Controller.GlobalCounter+10000]
	NextName := fmt.Sprintf(nextStudent.FirstName + " " + nextStudent.LastName)
	payload := entity.DashboardPayload{
		Name:            Name,
		StudentID:       student.StudentID,
		Faculty:         student.Faculty,
		Major:           student.Major,
		NextStudentName: strings.TrimSpace(NextName),
		Counter:         hl.Controller.GlobalCounter,
		Remaining:       len(hl.Controller.StudentList) - hl.Controller.GlobalCounter,
	}

	return e.JSON(http.StatusOK, payload)
}

func (hl handlers) IncrementCounter(e echo.Context) error {
	if hl.Controller.Mode != "sensor" {
		return e.JSON(http.StatusBadRequest, "Current Mode is not sensor")
	}
	hl.Controller.GlobalCounter += 1
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) DecrementCounter(e echo.Context) error {
	if hl.Controller.Mode != "sensor" {
		return e.JSON(http.StatusBadRequest, "Current Mode is not sensor")
	}
	if hl.Controller.GlobalCounter <= 0 {
		return e.JSON(http.StatusBadRequest, "Counter cannot be less than zero")
	}
	hl.Controller.GlobalCounter -= 1
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) SwitchMode(e echo.Context) error {
	mode := hl.Controller.Mode
	if mode == "auto" {
		hl.Controller.Mode = "sensor"
	} else {
		hl.Controller.Mode = "auto"
	}
	return e.JSON(http.StatusOK, hl.Controller.Mode)
}

// api to convert order of reading to counter of script
func (hl handlers) OrderToCounter(e echo.Context) error {
	orderOfReceiveParam := e.QueryParam("orderOfReceive")
	facultyParam := e.QueryParam("faculty")
	orderOfReceive, err := strconv.Atoi(orderOfReceiveParam)
	if err != nil {
		return e.JSON(http.StatusOK, err)
	}

	var found bool
	var filtered_script []entity.IndividualPayload
	filtered_script = append(filtered_script, entity.IndividualPayload{})
	for _, payload := range hl.Controller.Script {
		if payload.Type == "student name" {
			if payload.Data.(entity.StudentPayload).Faculty == facultyParam {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if payload.Type == "script" {
			if payload.Data.(entity.AnnouncerPayload).Faculty == facultyParam {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if found {
			if payload.Type == "student name" {
				if payload.Data.(entity.StudentPayload).Faculty != facultyParam {
					break
				}
			}
			if payload.Type == "script" {
				if payload.Data.(entity.AnnouncerPayload).Faculty != facultyParam {
					break
				}
			}
		}
	}
	filtered_script = append(filtered_script, entity.IndividualPayload{})
	//use orderOfReceive to find student in filtered_script
	counter := -1
	for i, payload := range filtered_script {
		if payload.Type == "student name" {
			if payload.Data.(entity.StudentPayload).OrderOfReading == orderOfReceive {
				counter = i
			}
		}
	}

	if counter == -1 {
		return e.JSON(http.StatusBadRequest, "Student not found")
	}

	//index previous entry to check if it is a script
	if filtered_script[counter-1].Type == "script" {
		counter -= 1
	}

	return e.JSON(http.StatusOK, counter)
}

func (hl handlers) GroupAnnouncersByFaculty(e echo.Context) error {
	faculties, err := hl.Controller.MySQLConn.QueryUniqueFaculties()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err)
	}

	studentsSlice := make([]entity.Student, 0, len(hl.Controller.StudentList))
	for _, student := range hl.Controller.StudentList {
		studentsSlice = append(studentsSlice, student)
	}
	sort.SliceStable(studentsSlice, func(i, j int) bool {
		return studentsSlice[i].OrderOfReceive < studentsSlice[j].OrderOfReceive
	})

	groupedAnnouncers := make(entity.AnnouncerGroupByFacultyPayload)

	for _, faculty := range faculties {
		for _, announcer := range hl.Controller.AnnouncerList {
			for _, student := range studentsSlice {
				if student.Faculty == faculty.Faculty && announcer.Start <= student.OrderOfReceive && announcer.End >= student.OrderOfReceive {
					if _, exists := groupedAnnouncers[faculty.Faculty]; !exists {
						groupedAnnouncers[faculty.Faculty] = []entity.AnnouncerGroupByFaculty{}
					}
					if !utility.AnnouncerAlreadyAdded(groupedAnnouncers[faculty.Faculty], announcer.AnnouncerID) {
						Counter, err := hl.Controller.OrderToCounter(student.OrderOfReceive, faculty.Faculty)
						if err != nil {
							continue
						}
						groupedAnnouncers[faculty.Faculty] = append(groupedAnnouncers[faculty.Faculty], entity.AnnouncerGroupByFaculty{
							AnnouncerID:   announcer.AnnouncerID,
							AnnouncerName: announcer.AnnouncerName,
							FirstOrder:    announcer.Start,
							LastOrder:     announcer.End,
							StartCounter:  Counter,
						})
					}
					break
				}
			}
		}
	}

	//sort announcer
	for faculty, announcers := range groupedAnnouncers {
		sort.Slice(announcers, func(i, j int) bool {
			return announcers[i].FirstOrder < announcers[j].FirstOrder
		})
		groupedAnnouncers[faculty] = announcers
	}
	// sort faculty back to original order
	orderedResponse := make(map[string][]entity.AnnouncerGroupByFaculty)
	for _, faculty := range faculties {
		if announcers, exists := groupedAnnouncers[faculty.Faculty]; exists {
			orderedResponse[faculty.Faculty] = announcers
		}
	}

	return e.JSON(http.StatusOK, orderedResponse)
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	// utility
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/api/ord-to-count", hl.OrderToCounter)
	e.GET("/api/switchMode", hl.SwitchMode)
	e.GET("/api/incrementCounter", hl.IncrementCounter)
	e.GET("/api/decrementCounter", hl.DecrementCounter)
	e.GET("/api/counter", hl.CounterAPI)
	e.GET("/api/dashboard", hl.DashboardAPI)
	e.GET("/api/practice/announce", hl.PracticeAnnounceAPI)

	// files and pages
	e.GET("/*", hl.Mainpage)
	e.Static("/assets", "html/dist/assets")

	//announcers
	e.GET("/api/announce", hl.AnnounceAPI)
	e.POST("/api/insert-announcer", hl.InsertAnnouncer)
	e.PUT("/api/update-announcer", hl.UpdateAnnouncer)
	e.GET("/api/announcers", hl.GetAnnouncers)
	e.GET("/api/grouped-announcers", hl.GroupAnnouncersByFaculty)
	e.DELETE("/api/delete-announcer", hl.DeleteAnnouncer)

	//students
	e.PUT("/api/students-list", hl.UpdateStudentList)
	e.PUT("/api/notes", hl.UpdateNotes)

	//faculty
	e.GET("/api/faculties", hl.GetFacultiesAPI)

	//testing
	e.GET("/test", hl.GroupAnnouncersByFaculty)

	//middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
}
