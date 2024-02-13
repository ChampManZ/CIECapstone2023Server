package utility

import (
	"capstone/server/entity"
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type MySQLDB struct {
	*sql.DB
}

func NewMySQLConn(dsn string) (*MySQLDB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &MySQLDB{db}, nil
}

func (db *MySQLDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.Query(query, args...)
}

func (db *MySQLDB) UpdateStudentList(jsonData []entity.StudentData) error {
	for _, student := range jsonData {
		studentID, err := strconv.Atoi(student.StudentID)
		if err != nil {
			return err
		}
		receiveOrder, err := strconv.Atoi(student.ReceiveOrder)
		if err != nil {
			return err
		}

		// Certificate
		certID, err := db.uinsertCertificate(student.CerName, student.FacultyName, student.CurrName, student.Honor)
		if err != nil {
			return err
		}

		// Student
		err = db.uinsertStudent(studentID, receiveOrder, student.Name, student.Surname, student.NameRead, int(certID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MySQLDB) uinsertCertificate(degree, faculty, major, honor string) (int64, error) {
	var certID int64
	// Check if exists
	query := `SELECT CertificateID FROM Certificate WHERE Degree = ? AND Faculty = ? AND Major = ? AND Honor = ?`
	err := db.QueryRow(query, degree, faculty, major, honor).Scan(&certID)
	if err == sql.ErrNoRows {
		// Insert if not exists
		insertQuery := `INSERT INTO Certificate(Degree, Faculty, Major, Honor) VALUES(?, ?, ?, ?)`
		res, err := db.Exec(insertQuery, degree, faculty, major, honor)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	} else if err != nil {
		return 0, err
	}
	// Return existing if found
	return certID, nil
}

func (db *MySQLDB) uinsertStudent(studentID, orderOfReceive int, firstname, surname, nameRead string, certificateID int) error {
	// Check exists
	var id int
	query := `SELECT StudentID FROM Student WHERE StudentID = ?`
	err := db.QueryRow(query, studentID).Scan(&id)
	if err == sql.ErrNoRows {
		// Insert if not exist
		insertQuery := `INSERT INTO Student(StudentID, OrderOfReceive, Firstname, Surname, NamePronunciation, CertificateID) VALUES(?, ?, ?, ?, ?, ?)`
		_, err := db.Exec(insertQuery, studentID, orderOfReceive, firstname, surname, nameRead, certificateID)
		return err
	} else if err != nil {
		return err
	} else {
		// Update if found
		updateQuery := `UPDATE Student SET OrderOfReceive = ?, Firstname = ?, Surname = ?, NamePronunciation = ?, CertificateID = ? WHERE StudentID = ?`
		_, err := db.Exec(updateQuery, orderOfReceive, firstname, surname, nameRead, certificateID, studentID)
		return err
	}
}

func (db *MySQLDB) QueryStudentsToMap() (map[int]entity.Student, error) {
	query := `
	SELECT 
		s.StudentID,
		s.OrderOfReceive, 
		s.Firstname, 
		s.Surname, 
		CONCAT(c.Faculty, ' ', c.Degree, 'สาขาวิชา', c.Major, ' ',
			CASE c.Honor 
				WHEN 0 THEN '' 
				WHEN 1 THEN 'เกียรตินิยมอันดับ 1' 
				WHEN 2 THEN 'เกียรตินิยมอันดับ 2' 
			END) AS Certificate, 
		COALESCE(nr.SavedNameRead, s.NamePronunciation) AS NameRead,
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
	LEFT JOIN 
		NameRead nr ON nr.NameReadStudentID = s.StudentID
	ORDER BY 
		s.OrderOfReceive ASC;
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make(map[int]entity.Student)
	var seenFaculty []string
	counter := 0
	for rows.Next() {
		var s entity.Student
		if err := rows.Scan(
			&s.StudentID,
			&s.OrderOfReceive,
			&s.FirstName,
			&s.LastName,
			&s.Certificate,
			&s.Reading,
			&s.RegReading,
			&s.Degree,
			&s.Faculty,
			&s.Major,
			&s.Honor); err != nil {
			return nil, err
		}
		major := s.Major
		if !IsFirstCharNotEnglish(s.Major) {
			major = fmt.Sprintf(" " + s.Major)
		}
		honor := s.Honor
		if honor != "" {
			honor = fmt.Sprintf(" " + s.Honor)
		}

		faculty := s.Faculty
		if IsNotInList(faculty, seenFaculty) {
			seenFaculty = append(seenFaculty, faculty)
			faculty = fmt.Sprintf(faculty + " ")
		} else {
			faculty = ""
		}

		s.Certificate = fmt.Sprintf(faculty + s.Degree + "สาขาวิชา" + major + honor)
		students[counter] = s
		counter++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (db *MySQLDB) QueryCounter() int {
	var currentValue int
	query := `SELECT CurrentValue FROM Counter WHERE ID = 1`
	err := db.QueryRow(query).Scan(&currentValue)

	if err != nil {
		if err == sql.ErrNoRows {
			insertQuery := `INSERT INTO Counter (ID, CurrentValue) VALUES (1, 0) ON DUPLICATE KEY UPDATE CurrentValue = 0`
			_, err := db.Exec(insertQuery)
			if err != nil {
				log.Fatalf("Failed to insert/update counter: %v", err)
			}
			return 0
		} else {
			log.Fatalf("Failed to query counter: %v", err)
		}
	}
	return currentValue
}

func (db *MySQLDB) QueryUniqueFaculties() ([]string, error) {
	query := `
    SELECT c.Faculty, s.OrderOfReceive
    FROM Certificate c
    JOIN Student s ON c.CertificateID = s.CertificateID
    ORDER BY s.OrderOfReceive ASC;
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facultyMap := make(map[string]bool)
	var faculties []string
	for rows.Next() {
		var faculty string
		var orderOfReceive int
		if err := rows.Scan(&faculty, &orderOfReceive); err != nil {
			return nil, err
		}
		if _, exists := facultyMap[faculty]; !exists {
			faculties = append(faculties, faculty)
			facultyMap[faculty] = true
		}
	}
	return faculties, nil
}

func (db *MySQLDB) UpdateAnnouncerQuery(announcerID int, announcerName, announcerScript, sessionOfAnnounce string, firstOrder, lastOrder int) error {
	query := `UPDATE Announcer 
	          SET AnnouncerName = ?, AnnouncerScript = ?, SessionOfAnnounce = ?, FirstOrder = ?, LastOrder = ?
	          WHERE AnnouncerID = ?`

	_, err := db.Exec(query, announcerName, announcerScript, sessionOfAnnounce, firstOrder, lastOrder, announcerID)
	if err != nil {
		return fmt.Errorf("failed to update announcer: %w", err)
	}

	return nil
}

func (db *MySQLDB) UpdateNote(orderOfReceive int, note string) error {
	var studentID int
	findIDQuery := `SELECT StudentID FROM Student WHERE OrderOfReceive = ?`
	err := db.QueryRow(findIDQuery, orderOfReceive).Scan(&studentID)
	if err != nil {
		return err
	}

	var exists bool
	checkExistenceQuery := `SELECT EXISTS(SELECT 1 FROM NameRead WHERE NameReadStudentID = ?)`
	err = db.QueryRow(checkExistenceQuery, studentID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		updateQuery := `UPDATE NameRead SET SavedNameRead = ? WHERE NameReadStudentID = ?`
		_, err = db.Exec(updateQuery, note, studentID)
		if err != nil {
			return err
		}
	} else {
		insertQuery := `INSERT INTO NameRead (NameReadStudentID, SavedNameRead) VALUES (?, ?)`
		_, err = db.Exec(insertQuery, studentID, note)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MySQLDB) QueryAnnouncers() (map[int]entity.Announcer, error) {
	query := `SELECT AnnouncerID, AnnouncerName, AnnouncerScript, SessionOfAnnounce, FirstOrder, LastOrder FROM Announcer`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying announcers: %w", err)
	}
	defer rows.Close()

	announcers := make(map[int]entity.Announcer)

	for rows.Next() {
		var a entity.Announcer
		if err := rows.Scan(&a.AnnouncerID, &a.AnnouncerName, &a.AnnouncerScript, &a.Session, &a.Start, &a.End); err != nil {
			return nil, fmt.Errorf("scanning announcer: %w", err)
		}
		announcers[a.AnnouncerID] = a
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over announcers: %w", err)
	}

	return announcers, nil
}

func (db *MySQLDB) InsertAnnouncer(announcerName, announcerScript, SessionOfAnnounce string, firstOrder, lastOrder int) error {
	query := `INSERT INTO Announcer (AnnouncerName, AnnouncerScript, SessionOfAnnounce, FirstOrder, LastOrder) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, announcerName, announcerScript, SessionOfAnnounce, firstOrder, lastOrder)
	if err != nil {
		return err
	}
	return nil

}

func (db *MySQLDB) DeleteAnnouncer(announcerID int) error {
	query := `DELETE FROM Announcer WHERE AnnouncerID = ?`
	_, err := db.Exec(query, announcerID)
	if err != nil {
		return err
	}
	return nil
}
