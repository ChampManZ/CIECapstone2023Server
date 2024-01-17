package conx

import (
	"capstone/server/entity"
	"capstone/server/utility"
)

type Controller struct {
	GlobalCounter        int
	MicrocontrollerAlive bool
	StudentList          map[int]entity.Student
	MySQLConn            *utility.MySQLDB
}

func NewController() Controller {
	return Controller{
		GlobalCounter:        0,
		MicrocontrollerAlive: false,
		StudentList:          make(map[int]entity.Student),
	}
}

func (c Controller) IncrementGlobalCounter() int {
	c.GlobalCounter++
	return c.GlobalCounter
}

func (c Controller) GetStudentByCounter(counter int) (*entity.Student, bool) {
	student, ok := c.StudentList[counter]
	return &student, ok
}
