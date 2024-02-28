package conx

import (
	"capstone/server/entity"
	"capstone/server/utility"
	"fmt"
	"sort"
	"strings"
)

type Controller struct {
	GlobalCounter        int
	MicrocontrollerAlive bool
	StudentList          map[int]entity.Student
	AnnouncerList        map[int]entity.Announcer
	MySQLConn            *utility.MySQLDB
	Script               []entity.IndividualPayload
	Mode                 string
}

func NewController() Controller {
	return Controller{
		GlobalCounter:        0,
		MicrocontrollerAlive: false,
		StudentList:          make(map[int]entity.Student),
		AnnouncerList:        make(map[int]entity.Announcer),
		Mode:                 "sensor",
	}
}

func (c *Controller) IncrementGlobalCounter() int {
	c.GlobalCounter++
	return c.GlobalCounter
}

func (c *Controller) GetStudentByCounter(counter int) (*entity.Student, bool) {
	student, ok := c.StudentList[counter]
	return &student, ok
}

func (c *Controller) GenerateSript() error {
	var payloads []entity.IndividualPayload
	var seenAnnouncers []int
	//get student and announcer lists
	students := c.StudentList
	announcers := c.AnnouncerList

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
	//loop through sorted list and construct script
	for i, student := range sortedStudents {
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

		//first case
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
		if student.Faculty != previousStudent.Faculty {
			payloads = append(payloads, entity.IndividualPayload{})
		}
		if announcerScript != "" {
			payloads = append(payloads, entity.IndividualPayload{
				Type: "script",
				Data: entity.AnnouncerPayload{
					AnnouncerID: announcerID,
					Script:      strings.TrimSpace(announcerScript),
					Faculty:     student.Faculty,
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
			},
		})
		previousStudent = student
	}
	payloads = append(payloads, entity.IndividualPayload{})
	c.Script = payloads
	return nil
}
