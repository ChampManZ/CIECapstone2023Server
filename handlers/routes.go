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

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type handlers struct {
	Controller *conx.Controller
	Echo       *echo.Echo
}

func NewHandlers(controller *conx.Controller) handlers {
	return handlers{
		Controller: controller,
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
	var response entity.AnnounceAPIPayload
	prevPayload := entity.IndividualPayload{}
	currPayload := entity.IndividualPayload{}
	nextPayload := entity.IndividualPayload{}
	next2Payload := entity.IndividualPayload{}
	next3Payload := entity.IndividualPayload{}
	index := hl.Controller.GlobalCounter

	script := hl.Controller.Script
	script = append(script, entity.IndividualPayload{})

	if index-1 >= 0 && index-1 < len(script) {
		prevPayload = script[index-1]
	}

	if index >= 0 && index < len(script) {
		currPayload = script[index]
	}

	if index+1 >= 0 && index+1 < len(script) {
		nextPayload = script[index+1]
	}

	if index+2 >= 0 && index+2 < len(script) {
		next2Payload = script[index+2]
	}

	if index+2 >= 0 && index+2 < len(script) {
		next2Payload = script[index+2]
	}

	if index+3 >= 0 && index+3 < len(script) {
		next3Payload = script[index+3]
	}

	if currPayload.Type == "student name" {
		response.Faculty = currPayload.Data.(entity.StudentPayload).Faculty
		response.Session = currPayload.Data.(entity.StudentPayload).Session
		response.CurrentNumber = currPayload.Data.(entity.StudentPayload).Order
		response.MaxNumber = currPayload.Data.(entity.StudentPayload).FacultyMax
	} else if currPayload.Type == "script" {
		response.Faculty = currPayload.Data.(entity.AnnouncerPayload).Faculty
		response.Session = currPayload.Data.(entity.AnnouncerPayload).Session
		for _, payload := range script[index:] {
			if payload.Type == "student name" {
				response.CurrentNumber = payload.Data.(entity.StudentPayload).Order
				response.MaxNumber = payload.Data.(entity.StudentPayload).FacultyMax
				break
			}
		}
	}

	response.Mode = hl.Controller.Mode
	if index != 0 {
		response.Blocks = append(response.Blocks, entity.IndividualPayload{})
		response.Blocks = append(response.Blocks, script[0:index-1]...)
	}

	response.Blocks = append(response.Blocks, []entity.IndividualPayload{prevPayload, currPayload, nextPayload, next2Payload, next3Payload}...)

	for i := range response.Blocks {
		response.Blocks[i].BlockID = i
	}

	return e.JSON(200, response)
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
	filtered_script = append([]entity.IndividualPayload{{}}, filtered_script...)
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
		return e.JSON(http.StatusInternalServerError, "Student not found")
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
		err = hl.Controller.MySQLConn.UpdateAnnouncerQuery(tx, updateRequest.AnnouncerID, updateRequest.AnnouncerName, updateRequest.AnnouncerScript, updateRequest.Session, updateRequest.Start, updateRequest.End, updateRequest.IsBreak)
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
		err = hl.Controller.MySQLConn.InsertAnnouncer(tx, insertRequest.AnnouncerName, insertRequest.AnnouncerScript, insertRequest.Session, insertRequest.Start, insertRequest.End, insertRequest.IsBreak)
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
	response := hl.Controller.PrepareDashboardMQTT()
	return e.JSON(http.StatusOK, response)
}

func (hl handlers) IncrementCounter(e echo.Context) error {
	flag := e.QueryParam("client")
	if flag == "false" && hl.Controller.Script[hl.Controller.GlobalCounter].Type == "script" {
		hl.Controller.MqttClient.Publish("signal", 2, false, "3")
		return e.JSON(http.StatusOK, "OK")
	}
	hl.Controller.MqttClient.Publish("signal", 2, false, "1")

	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) DecrementCounter(e echo.Context) error {
	if hl.Controller.GlobalCounter <= 0 {
		return e.JSON(http.StatusBadRequest, "Counter cannot be less than zero")
	}

	hl.Controller.MqttClient.Publish("signal", 2, false, "2")
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) SwitchMode(e echo.Context) error {
	modeParam := e.QueryParam("mode")
	speedParam := e.QueryParam("speed")

	if modeParam != "auto" && modeParam != "sensor" {
		return e.JSON(http.StatusBadRequest, "Invalid mode parameter")
	}
	hl.Controller.Lock.Lock()
	switch modeParam {
	case "auto":
		hl.Controller.Mode = "auto"
	case "sensor":
		hl.Controller.Mode = "sensor"
	}
	hl.Controller.Lock.Unlock()

	hl.Controller.ModeChangeSig <- hl.Controller.Mode
	if speedParam != "" {
		speed, err := strconv.Atoi(speedParam)
		if err != nil {
			return e.JSON(http.StatusBadRequest, "Invalid speed parameter")
		}
		if speed < 0 {
			return e.JSON(http.StatusBadRequest, "Speed cannot be less than zero")
		}
		err = hl.Controller.AdjustAutoSpeed(speed)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, err.Error())
		}
	}

	data := entity.ModeData{
		Mode:      hl.Controller.Mode,
		AutoSpeed: hl.Controller.AutoSpeed,
	}

	return e.JSON(http.StatusOK, data)
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

func (hl handlers) GetCounter(e echo.Context) error {
	return e.JSON(http.StatusOK, hl.Controller.GlobalCounter)
}

func (hl handlers) ResetCounter(e echo.Context) error {
	sessionParam := e.QueryParam("session")
	var Index int
	if sessionParam == "" {
		return e.JSON(http.StatusBadRequest, "Session parameter not found")
	}

	if sessionParam == "morning" {
		Index = 0
	} else if sessionParam == "afternoon" {
		Index = 0 // Default to 0 if not found
		for i, payload := range hl.Controller.Script {
			if payload.Type == "student name" {
				// Assuming Data is of type StudentPayload and has a Session field
				studentData, ok := payload.Data.(entity.StudentPayload)
				if ok && studentData.Session == "เช้า" {
					Index = i + 1
				}
			}
		}
	}

	//set current and next student
	var currStudent, nextStudent entity.StudentPayload
	var found bool
	for _, payload := range hl.Controller.Script[Index:] {

		if payload.Type != "student name" {
			continue
		}

		if !found {
			currStudent = payload.Data.(entity.StudentPayload)
			found = true
			continue
		}
		if found {
			nextStudent = payload.Data.(entity.StudentPayload)
			break
		}

	}

	hl.Controller.CurrentStudent = &currStudent
	hl.Controller.NextStudent = &nextStudent

	hl.Controller.Lock.Lock()
	defer hl.Controller.Lock.Unlock()
	hl.Controller.GlobalCounter = Index
	hl.Controller.MqttClient.Publish("signal", 2, false, "4")
	return e.JSON(http.StatusOK, "OK")
}

func (hl handlers) AdjustAutoSpeed(e echo.Context) error {
	speedParam := e.QueryParam("speed")
	if speedParam == "" {
		return e.JSON(http.StatusBadRequest, "Speed parameter not found")
	}
	speed, err := strconv.Atoi(speedParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid speed parameter")
	}
	if speed < 0 {
		return e.JSON(http.StatusBadRequest, "Speed cannot be less than zero")
	}

	// Safely update AutoSpeed
	hl.Controller.Lock.Lock()
	hl.Controller.AutoSpeed = speed
	hl.Controller.Lock.Unlock()

	select {
	case hl.Controller.SpeedChangeSig <- speed:
		// Speed update signal sent successfully
	default:
		// Prevent chan block if signal sent is in progress
	}

	return e.JSON(http.StatusOK, map[string]string{"new_speed": speedParam})

}

func (hl handlers) SetCounter(e echo.Context) error {
	CounterParam := e.QueryParam("counter")
	if CounterParam == "" {
		return e.JSON(http.StatusBadRequest, "Counter parameter not found")
	}

	Counter, err := strconv.Atoi(CounterParam)
	if err != nil {
		return e.JSON(http.StatusBadRequest, "Invalid counter parameter")
	}

	hl.Controller.Lock.Lock()
	defer hl.Controller.Lock.Unlock()
	hl.Controller.GlobalCounter = Counter
	return e.JSON(http.StatusOK, Counter)
}

func (hl handlers) RegisterRoutes(e *echo.Echo) {
	// utility
	e.GET("/healthcheck", hl.Healthcheck)
	e.GET("/api/ord-to-count", hl.OrderToCounter)
	e.GET("/api/switchMode", hl.SwitchMode)
	e.GET("/api/dashboard", hl.DashboardAPI)

	//Display
	e.GET("/api/announce", hl.AnnounceAPI)
	e.GET("/api/practice/announce", hl.PracticeAnnounceAPI)

	//Counter
	e.GET("/api/incrementCounter", hl.IncrementCounter)
	e.GET("/api/decrementCounter", hl.DecrementCounter)
	e.GET("/api/getCounter", hl.GetCounter)
	e.GET("/api/resetCounter", hl.ResetCounter)
	e.GET("/api/counter", hl.CounterAPI)
	e.GET("/api/autoSpeed", hl.AdjustAutoSpeed)
	e.GET("/api/setCounter", hl.SetCounter)

	// files and pages
	e.GET("/*", hl.Mainpage)
	e.Static("/assets", "html/dist/assets")

	//announcers
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
	e.GET("/test", hl.TestscriptAPI)

	//middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
}
