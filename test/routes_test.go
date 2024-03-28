package test

import (
	conx "capstone/server/controller"
	"capstone/server/entity"
	"capstone/server/handlers"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// The Healthcheck API should return a 200 status code and the string "OK".
func TestHealthcheckAPI(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Check the response body
	expectedBody := "OK"
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected response body %s, but got %s", expectedBody, rec.Body.String())
	}

	c.String(http.StatusOK, "OK")
}

// The Mainpage API should return a file named "index.html".
func TestMainpageAPI(t *testing.T) {
	// Initialize the echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Invoke the code under test
	hl := handlers.NewHandlers(conx.NewController())
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
	// Initialize Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/announce", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	// Mock the controller and student list
	mockController := new(conx.Controller) // Use the fully qualified name of MockController
	mockStudent := entity.Student{
		OrderOfReceive: 1,
		FirstName:      "John",
		LastName:       "Doe",
		Notes:          "Some notes",
		Certificate:    "Certificate",
	}
	mockController.StudentList = map[int]entity.Student{1: mockStudent}

	// Invoke the code under test
	hl := handlers.NewHandlers(conx.NewController())
	err := hl.AnnounceAPI(c)
	assert.NoError(t, err)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check the response body
	var responseBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	assert.NoError(t, err)

	// Assert type is either "student name" or "script"
	assert.Contains(t, []string{"student name", "script"}, responseBody["type"])

	// Check keys and data types in the "data" field
	data, ok := responseBody["data"].(map[string]interface{})
	assert.True(t, ok, "Expected 'data' field to be a map[string]interface{}")

	switch responseBody["type"] {
	case "student name":
		assert.Equal(t, 1, data["OrderOfReceive"])
		assert.IsType(t, "", data["FirstName"])
		assert.IsType(t, "", data["LastName"])
		assert.IsType(t, "", data["Notes"])
		assert.IsType(t, "", data["Certificate"])
	case "script":
		assert.IsType(t, 1, data["announcer_id"])
		assert.IsType(t, "", data["script"])
	}

	// Check the "block_id" field
	assert.IsType(t, 1, responseBody["block_id"])
}

func TestPracticeAnnounceAPI(t *testing.T) {
	// Initialize Echo context
	e := echo.New()

	// Mock the controller and student list
	mockController := new(conx.Controller)
	mockStudent := entity.Student{
		OrderOfReceive: 1,
		FirstName:      "John",
		LastName:       "Doe",
		Notes:          "Some notes",
		Certificate:    "Certificate",
	}
	mockController.StudentList = map[int]entity.Student{1: mockStudent}

	// Iterate through test cases
	for _, tc := range []struct {
		Name                 string
		QueryParams          map[string]string
		ExpectedStatusCode   int
		ExpectedResponseBody string
	}{
		{
			Name: "StartGreaterThanTotalStudents",
			QueryParams: map[string]string{
				"start":   "10",
				"amount":  "1",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusOK,
			ExpectedResponseBody: "[]",
		},
		{
			Name: "StartLessThanTotalStudents",
			QueryParams: map[string]string{
				"start":   "2",
				"amount":  "1",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusOK,
			ExpectedResponseBody: `[{"Type":"student name","Data":{"OrderOfReading":1,"Name":"John Doe","Reading":"Some notes","Certificate":"Certificate"}}]`,
		},
		{
			Name: "InvalidStartParameter",
			QueryParams: map[string]string{
				"start":   "abc",
				"amount":  "1",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: `"Invalid start parameter"`,
		},
		{
			Name: "InvalidStartNegativeValue",
			QueryParams: map[string]string{
				"start":   "-2",
				"amount":  "1",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: `"Invalid start parameter"`,
		},
		{
			Name: "InvalidAmountParameter",
			QueryParams: map[string]string{
				"start":   "1",
				"amount":  "abc",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: `"Invalid amount parameter"`,
		},
		{
			Name: "InvalidAmountNegativeValue",
			QueryParams: map[string]string{
				"start":   "1",
				"amount":  "-2",
				"faculty": "Computer Science",
			},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: `"Invalid amount parameter"`,
		},
		{
			Name: "EmptyFaculty",
			QueryParams: map[string]string{
				"start":   "1",
				"amount":  "1",
				"faculty": "",
			},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: `"Faculty must be provided as a non-empty string"`,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			// Prepare the request
			query := ""
			for key, value := range tc.QueryParams {
				query += key + "=" + value + "&"
			}
			req := httptest.NewRequest(http.MethodGet, "/api/practice-announce?"+query[:len(query)-1], nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Invoke the code under test
			hl := handlers.NewHandlers(conx.NewController())
			err := hl.PracticeAnnounceAPI(c)
			assert.NoError(t, err)

			// Check the response status code
			assert.Equal(t, tc.ExpectedStatusCode, rec.Code)

			// Check the response body
			assert.JSONEq(t, tc.ExpectedResponseBody, rec.Body.String())
		})
	}
}
