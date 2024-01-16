package entity

//TO DO: refactor
type Payload struct {
	Previous interface{} `json:"prev"`
	Current  interface{} `json:"curr"`
	Next     interface{} `json:"next"`
}

type IndividualPayload struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type StudentPayload struct {
	OrderOfReading int    `json:"OoR"`
	Name           string `json:"name"`
	Reading        string `json:"reading"`
	Note           string `json:"note"`
	Certificate    string `json:"cert"`
}

type Student struct {
	StudentID   int
	FirstName   string
	LastName    string
	Certificate string
	Notes       string
}
