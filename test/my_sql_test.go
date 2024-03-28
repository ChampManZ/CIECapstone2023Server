package test

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"capstone/server/utility"
	"database/sql"
	"reflect"
	"testing"
)

// Connect to MySQL database successfully
func TestConnectToMySQLSuccessfully(t *testing.T) {
	// Mock the sql.Open function
	sql.Open("181314a4c00344d5a62cfc8c2059208e.s1.eu.hivemq.cloud", "capstone")

	// Call the code under test
	_, err := utility.NewMySQLConn("dsn")

	// Assert that there is no error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

// Fail to connect to MySQL database
func TestFailToConnectToMySQL(t *testing.T) {
	// Mock the sql.Open function
	sql.Open("181314a4c00344d5a62cfc8c2059208e.s1.eu.hivemq.cloud", "capstone")

	// Call the code under test
	_, err := utility.NewMySQLConn("dsn")

	// Assert that there is an error
	if err == nil {
		t.Error("Expected an error, but got nil")
	}
}

// Query students and map them to a map[int]entity.Student
func TestQueryStudentsToMap(t *testing.T) {
	// Mock the sql.DB and sql.Rows
	mockDB := &sql.DB{}
	mockRows := &sql.Rows{}
	mockDB.Query(`
	SELECT 
        s.StudentID,
        s.OrderOfReceive, 
        s.Firstname, 
        s.Surname, 
        CONCAT(c.Faculty,' ',c.Degree,'สาขาวิชา',c.Major,' ',
            CASE c.Honor 
                WHEN 0 THEN '' 
                WHEN 1 THEN 'เกียรตินิยมอันดับ 1' 
                WHEN 2 THEN 'เกียรตินิยมอันดับ 2' 
            END) AS Certificate, 
        s.NamePronunciation,
        c.Degree,
        c.Faculty,
        c.Major,
        CASE c.Honor 
            WHEN 0 THEN '' 
            WHEN 1 THEN 'เกียรตินิยมอันดับ 1' 
            WHEN 2 THEN 'เกียรตินิยมอันดับ 2' 
        END AS Honor
    FROM 
        Student s
    JOIN 
        Certificate c ON s.CertificateID = c.CertificateID
    ORDER BY 
        s.OrderOfReceive ASC;
	`)

	// Create a mock student entity
	mockStudent := entity.Student{
		StudentID:      1,
		OrderOfReceive: 1,
		FirstName:      "John",
		LastName:       "Doe",
		Certificate:    "Certificate",
		Notes:          "Notes",
		Degree:         "Degree",
		Faculty:        "Faculty",
		Major:          "Major",
		Honor:          "Honor",
	}

	// Mock the rows.Next function
	mockRows.Next()

	// Mock the rows.Scan function
	mockRows.Scan()

	// Call the code under test
	db := &utility.MySQLDB{DB: mockDB}
	students, err := db.QueryStudentsToMap()

	// Assert that there is no error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Assert that the students map is not empty
	if len(students) == 0 {
		t.Errorf("Expected non-empty students map, but got empty")
	}

	// Assert that the student entity is correctly mapped
	expectedStudent := mockStudent
	actualStudent := students[0]
	if !reflect.DeepEqual(expectedStudent, actualStudent) {
		t.Errorf("Expected student %v, but got %v", expectedStudent, actualStudent)
	}
}

// Return map[int]entity.Student and no error
func TestQueryStudentsToMapReturn(t *testing.T) {
	// Mock the sql.DB and sql.Rows
	mockDB := &sql.DB{}
	mockRows := &sql.Rows{}
	mockDB.Query(`
	SELECT 
        s.StudentID,
        s.OrderOfReceive, 
        s.Firstname, 
        s.Surname, 
        CONCAT(c.Faculty,' ',c.Degree,'สาขาวิชา',c.Major,' ',
            CASE c.Honor 
                WHEN 0 THEN '' 
                WHEN 1 THEN 'เกียรตินิยมอันดับ 1' 
                WHEN 2 THEN 'เกียรตินิยมอันดับ 2' 
            END) AS Certificate, 
        s.NamePronunciation,
        c.Degree,
        c.Faculty,
        c.Major,
        CASE c.Honor 
            WHEN 0 THEN '' 
            WHEN 1 THEN 'เกียรตินิยมอันดับ 1' 
            WHEN 2 THEN 'เกียรตินิยมอันดับ 2' 
        END AS Honor
    FROM 
        Student s
    JOIN 
        Certificate c ON s.CertificateID = c.CertificateID
    ORDER BY 
        s.OrderOfReceive ASC;
	`)

	// Call the code under test
	db := &utility.MySQLDB{DB: mockDB}
	students, err := db.QueryStudentsToMap()

	// Mock the rows.Next function
	mockRows.Next()

	// Mock the rows.Scan function
	mockRows.Scan()

	// Assert that there is no error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Assert that the students map is not empty
	if len(students) == 0 {
		t.Errorf("Expected non-empty students map, but got empty")
	}
}

// Query counter successfully
func TestQueryCounter(t *testing.T) {
	// Mock the sql.DB and sql.Row
	mockDB := &sql.DB{}
	mockRow := &sql.Row{}
	mockDB.Query(`
	SELECT c.Faculty, s.OrderOfReceive
    FROM Certificate c
    JOIN Student s ON c.CertificateID = s.CertificateID
    ORDER BY s.OrderOfReceive ASC;
	`)

	// Mock the row.Scan function
	mockRow.Scan(conx.NewController().GlobalCounter)

	// Call the code under test
	db := &utility.MySQLDB{DB: mockDB}
	currentValue := db.QueryCounter()

	// Assert that the current value is correct
	expectedValue := 10
	if currentValue != expectedValue {
		t.Errorf("Expected current value %d, but got %d", expectedValue, currentValue)
	}
}
