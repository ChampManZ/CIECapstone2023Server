package entity

type AnnouncePayload struct {
	Previous interface{} `json:"prev"`
	Current  interface{} `json:"curr"`
	Next     interface{} `json:"next"`
}

type CounterPayload struct {
	Current   interface{} `json:"curr"`
	Remaining interface{} `json:"remaining"`
}

type IndividualPayload struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type StudentPayload struct {
	OrderOfReading int    `json:"OoR"`
	Name           string `json:"name"`
	Reading        string `json:"reading"`
	RegReading     string `json:"reg_reading"`
	Faculty        string
	Certificate    string `json:"cert"`
}

type Student struct {
	StudentID      int
	FirstName      string
	LastName       string
	Certificate    string
	Reading        string
	RegReading     string
	OrderOfReceive int
	Degree         string
	Faculty        string
	Major          string
	Honor          string
}

type Announcer struct {
	AnnouncerID     int    `json:"AnnouncerID"`
	AnnouncerName   string `json:"AnnouncerName"`
	AnnouncerScript string `json:"AnnouncerScript"`
	Session         string `json:"SessionOfAnnounce"`
	Start           int    `json:"FirstOrder"`
	End             int    `json:"LastOrder"`
}

type AnnouncerPayload struct {
	AnnouncerID int    `json:"announcer_id"`
	Script      string `json:"script"`
	Faculty     string
}

// Student Data Struct from Query
type StudentData struct {
	ReceiveOrder string `json:"receive_order"`
	StudentID    string `json:"student_id"`
	Name         string `json:"name"`
	Surname      string `json:"surname"`
	NameRead     string `json:"name_read"`
	FacultyName  string `json:"faculty_name"`
	CurrName     string `json:"curr_name"`
	CerName      string `json:"cer_name"`
	Honor        string `json:"honor"`
}

type FacultySession struct {
	Faculty           string `json:"faculty"`
	SessionOfAnnounce string `json:"session"`
}

type DashboardPayload struct {
	Name            string `json:"name"`
	StudentID       int    `json:"studentID"`
	Faculty         string `json:"faculty"`
	Major           string `json:"major"`
	Counter         int    `json:"order"`
	Remaining       int    `json:"remain"`
	NextStudentName string `json:"nextName"`
}

type FacultySessionPayload struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

type AnnouncerGroupByFaculty struct {
	AnnouncerID   int    `json:"AnnouncerID"`
	AnnouncerName string `json:"AnnouncerName"`
	FirstOrder    int    `json:"FirstOrder"`
	LastOrder     int    `json:"LastOrder"`
	StartCounter  int    `json:"StartCounter"`
}

type AnnouncerGroupByFacultyPayload map[string][]AnnouncerGroupByFaculty
