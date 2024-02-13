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
	return e.File("html/main/index.html")
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
	announcers := hl.Controller.AnnouncerList
	var seenAnnouncers []int
	// if start-1 > 0 {
	// 	val, ok := sortedStudents[start-1].(entity.Student)
	// 	if ok {
	// 		previousStudent = &val
	// 	}
	// }

	for i, student := range sortedStudents {

		if i >= start && i < start+amount {
			if student == nil {
				payloads = append(payloads, entity.IndividualPayload{})
				continue
			}
			student := student.(entity.Student)

			announcerScript := ""
			announcerID := 0
			seen := false
			diff := false

			for _, announcer := range announcers {
				announcerID = announcer.AnnouncerID
				for _, item := range seenAnnouncers {
					if item == announcer.AnnouncerID {
						seen = true
					}
				}
				if !seen && student.OrderOfReceive == announcer.Start && student.OrderOfReceive <= announcer.End {
					announcerScript = announcer.AnnouncerScript
					seenAnnouncers = append(seenAnnouncers, announcer.AnnouncerID)
					break
				}
			}

			var certificateValue string
			if previousStudent != nil {
				degree := student.Degree
				if utility.IsFirstCharNotEnglish(degree) {
					degree = fmt.Sprintf("ปริญญา" + strings.TrimSpace(degree))
				} else {
					degree = fmt.Sprintf("ปริญญา " + strings.TrimSpace(degree))
				}
				if student.Degree != previousStudent.Degree {
					diff = true
					//certificateValue = fmt.Sprintf(certificateValue + strings.TrimSpace(degree))
					announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(degree))
				}
				major := student.Major
				if utility.IsFirstCharNotEnglish(major) {
					major = fmt.Sprintf("สาขาวิชา" + strings.TrimSpace(major))
				} else {
					major = fmt.Sprintf("สาขาวิชา " + strings.TrimSpace(major))
				}
				if student.Major != previousStudent.Major {
					diff = true
					//certificateValue = fmt.Sprintf(certificateValue + " " + strings.TrimSpace(major))
					announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(major))
				}
				if student.Honor != previousStudent.Honor {
					if previousStudent.Honor != "เกียรตินิยมอันดับ 2" {
						certificateValue = fmt.Sprintf(certificateValue + " " + student.Honor)
					} else {
						certificateValue = fmt.Sprintf(certificateValue + " " + strings.TrimSpace(major))
					}
				}
				if !(student.Major != previousStudent.Major ||
					student.Degree != previousStudent.Degree ||
					student.Honor != previousStudent.Honor) {
					certificateValue = ""
				}
				// if student.Major != previousStudent.Major ||
				// 	student.Degree != previousStudent.Degree ||
				// 	student.Honor != previousStudent.Honor {
				// 	certificateValue = student.Certificate
				// } else {
				// 	certificateValue = ""
				// }
			} else {
				degree := student.Degree
				if utility.IsFirstCharNotEnglish(degree) {
					degree = fmt.Sprintf("ปริญญา" + strings.TrimSpace(degree))
				} else {
					degree = fmt.Sprintf("ปริญญา " + strings.TrimSpace(degree))
				}
				major := student.Major
				if utility.IsFirstCharNotEnglish(major) {
					major = fmt.Sprintf("สาขาวิชา" + strings.TrimSpace(major))
				} else {
					major = fmt.Sprintf("สาขาวิชา " + strings.TrimSpace(major))
				}
				//certificateValue = fmt.Sprintf(certificateValue + strings.TrimSpace(degree))
				//certificateValue = fmt.Sprintf(certificateValue + " " + strings.TrimSpace(major))
				certificateValue = fmt.Sprintf(certificateValue + " " + strings.TrimSpace(student.Honor))
				announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(degree))
				announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(major))
			}

			certificateValue = strings.TrimSpace(certificateValue)
			announcerScript = strings.TrimSpace(announcerScript)
			if announcerScript != "" || diff {
				payload := entity.IndividualPayload{
					Type: "script",
					Data: entity.AnnouncerPayload{
						AnnouncerID: announcerID,
						Script:      announcerScript,
					},
				}
				payloads = append(payloads, payload)
			}

			payload := entity.IndividualPayload{
				Type: "student name",
				Data: entity.StudentPayload{
					OrderOfReading: student.OrderOfReceive,
					Name:           student.FirstName + " " + student.LastName,
					Reading:        student.Reading,
					RegReading:     student.RegReading,
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
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) GetFacultiesAPI(e echo.Context) error {
	faculties, err := hl.Controller.MySQLConn.QueryUniqueFaculties()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}
	return e.JSON(http.StatusOK, faculties)
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

	return e.JSON(http.StatusOK, "OK")

}

func (hl handlers) UpdateAnnouncer(e echo.Context) error {
	announcerIDParam := e.QueryParam("announcerID")
	announcerName := e.QueryParam("AnnouncerName")
	announcerScript := e.QueryParam("AnnouncerScript")
	SessionOfAnnounce := e.QueryParam("SessionOfAnnounce")
	firstOrderStr := e.QueryParam("FirstOrder")
	lastOrderStr := e.QueryParam("LastOrder")

	firstOrder, err := strconv.Atoi(firstOrderStr)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid FirstOrder")
	}

	lastOrder, err := strconv.Atoi(lastOrderStr)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid LastOrder")
	}

	announcerID, err := strconv.Atoi(announcerIDParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid announcerID parameter")
	}

	err = hl.Controller.MySQLConn.UpdateAnnouncerQuery(announcerID, announcerName, announcerScript, SessionOfAnnounce, firstOrder, lastOrder)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.String(http.StatusOK, "OK")
}

func (hl handlers) InsertAnnouncer(e echo.Context) error {
	announcerName := e.QueryParam("AnnouncerName")
	announcerScript := e.QueryParam("AnnouncerScript")
	SessionOfAnnounce := e.QueryParam("SessionOfAnnounce")
	firstOrderStr := e.QueryParam("FirstOrder")
	lastOrderStr := e.QueryParam("LastOrder")

	firstOrder, err := strconv.Atoi(firstOrderStr)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid FirstOrder")
	}

	lastOrder, err := strconv.Atoi(lastOrderStr)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid LastOrder")
	}

	err = hl.Controller.MySQLConn.InsertAnnouncer(announcerName, announcerScript, SessionOfAnnounce, firstOrder, lastOrder)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

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
	announcerIDParam := e.QueryParam("AnnouncerID")

	announcerID, err := strconv.Atoi(announcerIDParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid announcerID")
	}

	err = hl.Controller.MySQLConn.DeleteAnnouncer(announcerID)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	hl.Controller.AnnouncerList, err = hl.Controller.MySQLConn.QueryAnnouncers()
	if err != nil {
		return e.JSON(http.StatusInternalServerError, err.Error())
	}

	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/", hl.Mainpage)
	e.GET("/api/announce", hl.AnnounceAPI)
	e.GET("/api/counter", hl.CounterAPI)
	e.GET("/api/practice/announce", hl.PracticeAnnounceAPI)
	e.GET("/api/faculties", hl.GetFacultiesAPI)
	e.PUT("/api/notes", hl.UpdateNotes)
	e.PUT("/api/students-list", hl.UpdateStudentList)
	e.POST("/api/insert-announcer", hl.InsertAnnouncer)
	e.PUT("/api/update-announcer", hl.UpdateAnnouncer)
	e.GET("/api/announcers", hl.GetAnnouncers)
	e.DELETE("/api/delete-announcer", hl.DeleteAnnouncer)
	e.Static("/assets", "html/main/assets")
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
}
