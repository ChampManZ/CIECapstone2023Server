
	// Connect to MySQL database successfully
	func TestConnectToMySQLSuccessfully(t *testing.T) {
		// Mock the sql.Open function
		mockDB := &sql.DB{}
		mockOpen := func(driverName, dataSourceName string) (*sql.DB, error) {
			return mockDB, nil
		}
		sql.Open = mockOpen

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
		mockOpen := func(driverName, dataSourceName string) (*sql.DB, error) {
			return nil, errors.New("failed to connect")
		}
		sql.Open = mockOpen

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
		mockQuery := func(query string, args ...interface{}) (*sql.Rows, error) {
			return mockRows, nil
		}
		mockDB.Query = mockQuery

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
		mockNext := func() bool {
			return true
		}
		mockRows.Next = mockNext

		// Mock the rows.Scan function
		mockScan := func(dest ...interface{}) error {
			s := dest[0].(*entity.Student)
			*s = mockStudent
			return nil
		}
		mockRows.Scan = mockScan

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
		mockQuery := func(query string, args ...interface{}) (*sql.Rows, error) {
			return mockRows, nil
		}
		mockDB.Query = mockQuery

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
		mockNext := func() bool {
			return true
		}
		mockRows.Next = mockNext

		// Mock the rows.Scan function
		mockScan := func(dest ...interface{}) error {
			s := dest[0].(*entity.Student)
			*s = mockStudent
			return nil
		}
		mockRows.Scan = mockScan

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
	}

	// Query counter successfully
	func TestQueryCounter(t *testing.T) {
		// Mock the sql.DB and sql.Row
		mockDB := &sql.DB{}
		mockRow := &sql.Row{}
		mockQueryRow := func(query string, args ...interface{}) *sql.Row {
			return mockRow
		}
		mockDB.QueryRow = mockQueryRow

		// Mock the row.Scan function
		mockScan := func(dest ...interface{}) error {
			currentValue := dest[0].(*int)
			*currentValue = 10
			return nil
		}
		mockRow.Scan = mockScan

		// Call the code under test
		db := &utility.MySQLDB{DB: mockDB}
		currentValue := db.QueryCounter()

		// Assert that the current value is correct
		expectedValue := 10
		if currentValue != expectedValue {
			t.Errorf("Expected current value %d, but got %d", expectedValue, currentValue)
		}
	}
	