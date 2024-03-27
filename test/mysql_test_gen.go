package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryStudentsToMap(t *testing.T) {
	// Create a mock MySQLDB instance
	db := &MySQLDB{}

	// Mock the query result
	mockRows := &MockRows{}
	mockRows.On("Close").Return(nil)
	mockRows.On("Next").Return(true).Times(2).Return(false)
	mockRows.On("Scan", mock.AnythingOfType("*int"), mock.AnythingOfType("*int"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(nil)

	// Mock the db.Query method
	db.On("Query", mock.AnythingOfType("string")).Return(mockRows, nil)

	// Call the function being tested
	students, err := db.QueryStudentsToMap()

	// Assert that the function returned the expected result
	assert.NoError(t, err)
	assert.NotNil(t, students)
	assert.Equal(t, 2, len(students))

	// Assert any other expectations on the mock objects if needed
	mockRows.AssertExpectations(t)
	db.AssertExpectations(t)
}
