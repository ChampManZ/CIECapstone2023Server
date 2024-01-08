package entity

type Payload struct {
	Previous interface{}
	Current  interface{}
	Next     interface{}
}

type StudentPayload struct {
	OrderOfReading int
	Name           string
	Reading        string
	Note           string
}

type Student struct {
	StudentID   int
	FirstName   string
	LastName    string
	Certificate string
	Notes       string
}
