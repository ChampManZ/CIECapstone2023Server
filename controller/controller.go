package conx

import (
	"capstone/server/entity"
	"capstone/server/utility"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Controller struct {
	GlobalCounter        int
	MicrocontrollerAlive bool
	StudentList          map[int]entity.Student
	AnnouncerList        map[int]entity.Announcer
	MySQLConn            *utility.MySQLDB
	MqttClient           mqtt.Client
	Script               []entity.IndividualPayload
	Mode                 string
	AutoSpeed            int
	SpeedChangeSig       chan int
	ModeChangeSig        chan string
	stop                 chan struct{}
	Lock                 sync.Mutex
}

func NewController() Controller {
	return Controller{
		GlobalCounter:        0,
		MicrocontrollerAlive: false,
		StudentList:          make(map[int]entity.Student),
		AnnouncerList:        make(map[int]entity.Announcer),
		Mode:                 "sensor",
		ModeChangeSig:        make(chan string),
		SpeedChangeSig:       make(chan int, 1),
		stop:                 make(chan struct{}),
		AutoSpeed:            25,
	}
}

func (c *Controller) IncrementGlobalCounter() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.GlobalCounter++
	return c.GlobalCounter
}

func (c *Controller) DecrementGlobalCounter() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.GlobalCounter--
	return c.GlobalCounter
}

func (c *Controller) GetStudentByCounter(counter int) (*entity.Student, bool) {
	student, ok := c.StudentList[counter]
	return &student, ok
}

func (c *Controller) GenerateSript() error {
	var payloads []entity.IndividualPayload
	var seenAnnouncers []int
	var session string = "เช้า"
	//get student and announcer lists
	students := c.StudentList
	announcers := c.AnnouncerList

	facultyOrderCount := make(map[string]int)
	facultyMax := make(map[string]int)

	//count max for each faculty
	for _, student := range students {
		facultyMax[student.Faculty]++
	}

	//sort student list by order of receive
	keys := make([]int, 0, len(students))
	for k := range students {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	sortedStudents := make([]entity.Student, len(keys))
	for i, k := range keys {
		sortedStudents[i] = students[k]
	}
	var previousStudent entity.Student
	//payloads = append(payloads, entity.IndividualPayload{})
	//loop through sorted list and construct script
	for i, student := range sortedStudents {
		facultyOrderCount[student.Faculty]++
		var announcerID int = 0
		var announcerScript string = ""
		var certificateValue string = ""
		//check announcers
		for _, announcer := range announcers {
			//check if announcer not in seenAnnouncers
			seen := false
			for _, seenAnnouncer := range seenAnnouncers {
				if seenAnnouncer == announcer.AnnouncerID {
					seen = true
					break
				}
			}
			if seen {
				continue
			}

			if student.OrderOfReceive == announcer.Start {
				announcerScript = announcer.AnnouncerScript
				seenAnnouncers = append(seenAnnouncers, announcer.AnnouncerID)
				announcerID = announcer.AnnouncerID
				break
			}
		}

		order := facultyOrderCount[student.Faculty]
		max := facultyMax[student.Faculty]

		announcerScript, certificateValue = constructScript(i, student, announcerScript, previousStudent, certificateValue)
		// if student.Faculty != previousStudent.Faculty {
		// 	payloads = append(payloads, entity.IndividualPayload{})
		// }

		if announcers[announcerID].Session == "บ่าย" {
			session = "บ่าย"
		}

		if announcerScript != "" {
			payloads = append(payloads, entity.IndividualPayload{
				Type: "script",
				Data: entity.AnnouncerPayload{
					AnnouncerID: announcerID,
					Script:      strings.TrimSpace(announcerScript),
					Faculty:     student.Faculty,
					Session:     session,
				},
			})
		}
		payloads = append(payloads, entity.IndividualPayload{
			Type: "student name",
			Data: entity.StudentPayload{
				OrderOfReading: student.OrderOfReceive,
				Name:           student.FirstName + " " + student.LastName,
				Reading:        student.Reading,
				RegReading:     student.RegReading,
				Faculty:        student.Faculty,
				Certificate:    strings.TrimSpace(certificateValue),
				Session:        session,
				Order:          order,
				FacultyMax:     max,
			},
		})
		previousStudent = student
	}
	//payloads = append(payloads, entity.IndividualPayload{})
	c.Script = payloads
	return nil
}

func constructScript(i int, student entity.Student, announcerScript string, previousStudent entity.Student, certificateValue string) (string, string) {
	if i == 0 {
		degree := student.Degree
		if utility.IsFirstCharNotEnglish(degree) {
			degree = fmt.Sprintf("ปริญญา" + strings.TrimSpace(degree))
		} else {
			degree = fmt.Sprintf("ปริญญา " + strings.TrimSpace(degree))
		}
		announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(degree))

		major := student.Major
		if utility.IsFirstCharNotEnglish(major) {
			major = fmt.Sprintf("สาขาวิชา" + strings.TrimSpace(major))
		} else {
			major = fmt.Sprintf("สาขาวิชา " + strings.TrimSpace(major))
		}
		announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(major))
	} else {
		degree := student.Degree
		if utility.IsFirstCharNotEnglish(degree) {
			degree = fmt.Sprintf("ปริญญา" + strings.TrimSpace(degree))
		} else {
			degree = fmt.Sprintf("ปริญญา " + strings.TrimSpace(degree))
		}
		if student.Degree != previousStudent.Degree {
			announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(degree))
		}

		major := student.Major
		if utility.IsFirstCharNotEnglish(major) {
			major = fmt.Sprintf("สาขาวิชา" + strings.TrimSpace(major))
		} else {
			major = fmt.Sprintf("สาขาวิชา " + strings.TrimSpace(major))
		}
		if student.Major != previousStudent.Major {
			announcerScript = fmt.Sprintf(announcerScript + " " + strings.TrimSpace(major))
		}

		if student.Honor != previousStudent.Honor {
			if previousStudent.Honor != "เกียรตินิยมอันดับ 2" {
				certificateValue = fmt.Sprintf(certificateValue + " " + student.Honor)
			} else {
				certificateValue = fmt.Sprintf(certificateValue + " " + strings.TrimSpace(major))
			}
		}

	}
	return announcerScript, certificateValue
}

func (c *Controller) OrderToCounter(orderOfReceive int, faculty string) (int, error) {
	var found bool
	var filtered_script []entity.IndividualPayload
	filtered_script = append(filtered_script, entity.IndividualPayload{})
	for _, payload := range c.Script {
		if payload.Type == "student name" {
			if payload.Data.(entity.StudentPayload).Faculty == faculty {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if payload.Type == "script" {
			if payload.Data.(entity.AnnouncerPayload).Faculty == faculty {
				filtered_script = append(filtered_script, payload)
				found = true
			}
		}
		if found {
			if payload.Type == "student name" {
				if payload.Data.(entity.StudentPayload).Faculty != faculty {
					break
				}
			}
			if payload.Type == "script" {
				if payload.Data.(entity.AnnouncerPayload).Faculty != faculty {
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
		return -1, fmt.Errorf("student not found")
	}

	//index previous entry to check if it is a script
	if filtered_script[counter-1].Type == "script" {
		counter -= 1
	}

	return counter, nil
}

func (c *Controller) PublishMQTT() {
	var ticker *time.Ticker
	var tickerC <-chan time.Time

	for {
		select {
		case mode := <-c.ModeChangeSig:
			// Mode change handling remains the same
			if mode == "auto" && ticker == nil {
				// Initialize the ticker with the current AutoSpeed
				c.Lock.Lock()
				interval := time.Minute / time.Duration(c.AutoSpeed)
				c.Lock.Unlock()

				ticker = time.NewTicker(interval)
				tickerC = ticker.C
			} else if mode != "auto" && ticker != nil {
				ticker.Stop()
				ticker = nil
				tickerC = nil
			}
		case <-tickerC:
			// Perform MQTT publishing
			c.MqttClient.Publish("signal", 0, false, "1")
		case newSpeed := <-c.SpeedChangeSig:
			// Handle speed change
			if ticker != nil {
				ticker.Stop()
				interval := time.Minute / time.Duration(newSpeed)
				ticker = time.NewTicker(interval)
				tickerC = ticker.C
			}
		case <-c.stop:
			if ticker != nil {
				ticker.Stop()
			}
			return // Exit the goroutine
		}
	}

}
