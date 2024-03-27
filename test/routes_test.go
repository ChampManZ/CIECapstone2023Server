package test

import (
	"capstone/server/entity"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// The Healthcheck API should return a 200 status code and the string "OK".
func TestHealthcheckAPI(t *testing.T) {
	// Initialize the echo context
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the code under test
	hl := handlers{}
	err := hl.Healthcheck(c)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := "OK"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}

// The Mainpage API should return a file named "index.html".
func TestMainpageAPI(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the code under test
	hl := handlers{}
	err := hl.Mainpage(c)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := "index.html"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}

// The AnnounceAPI should return a JSON payload with the next student's name and information.
func TestAnnounceAPI(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/announce", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock the controller and student list
	mockController := &mocks.Controller{}
	mockStudent := entity.Student{
		OrderOfReceive: 1,
		FirstName:      "John",
		LastName:       "Doe",
		Notes:          "Some notes",
		Certificate:    "Certificate",
	}
	mockController.On("GetStudentList").Return(map[int]entity.Student{1: mockStudent}, nil)

	// Invoke the code under test
	hl := handlers{Controller: mockController}
	err := hl.AnnounceAPI(c)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := `{
			"Type": "student name",
			"Data": {
				"OrderOfReading": 1,
				"Name": "John Doe",
				"Reading": "Some notes",
				"Certificate": "Certificate"
			}
		}`
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}

// When the start parameter in PracticeAnnounceAPI is greater than the number of students, an empty list should be returned.
func TestPracticeAnnounceAPI_StartGreaterThanStudents(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/practice/announce?start=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock the controller and student list
	mockController := &mocks.Controller{}
	mockController.On("GetStudentList").Return(map[int]entity.Student{}, nil)

	// Invoke the code under test
	hl := handlers{Controller: mockController}
	err := hl.PracticeAnnounceAPI(c)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := "[]"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}

// When the amount parameter in PracticeAnnounceAPI is greater than the number of students, the list should be truncated to the number of students.
func TestPracticeAnnounceAPI_AmountGreaterThanStudents(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/practice/announce?amount=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock the controller and student list
	mockController := &mocks.Controller{}
	mockStudent := entity.Student{
		OrderOfReceive: 1,
		FirstName:      "John",
		LastName:       "Doe",
		Notes:          "Some notes",
		Certificate:    "Certificate",
	}
	mockController.On("GetStudentList").Return(map[int]entity.Student{1: mockStudent}, nil)

	// Invoke the code under test
	hl := handlers{Controller: mockController}
	err := hl.PracticeAnnounceAPI(c)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Check the response status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}

	// Check the response body
	expectedBody := `[{
			"Type": "student name",
			"Data": {
				"OrderOfReading": 1,
				"Name": "John Doe",
				"Reading": "Some notes",
				"Certificate": "Certificate"
			}
		}]`
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}

// When the orderOfReceive parameter in updateNotes API is not a valid integer, a 400 status code should be returned.
func TestUpdateNotesAPI_InvalidOrderOfReceive(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/notes?orderOfReceive=abc&note=SomeNote", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the code under test
	hl := handlers{}
	err := hl.updateNotes(c)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}

	// Check the response status code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, rec.Code)
	}

	// Check the response body
	expectedBody := "Invalid order of receive parameter"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}
}
