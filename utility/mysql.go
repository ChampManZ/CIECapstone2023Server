package utility

import (
	"capstone/server/entity"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type MySQLDB struct {
	*sql.DB
}

func NewMySQLConn(dsn string) (*MySQLDB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Println("Error connecting to MySQL:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Ping the database to check connection
		err = db.Ping()
		if err == nil {
			break
		}
		log.Println("Error pinging MySQL:", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to MySQL!")
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

func (db *MySQLDB) QueryUniqueFaculties() ([]entity.FacultySession, error) {
	// Check for the presence of any afternoon session first
	var firstOrderAfternoon int
	var afternoonFaculty string
	err := db.QueryRow(`
        SELECT MIN(FirstOrder)
        FROM Announcer a
        WHERE a.SessionOfAnnounce = 'บ่าย'
        GROUP BY a.AnnouncerID
        ORDER BY MIN(FirstOrder) ASC
        LIMIT 1`).Scan(&firstOrderAfternoon)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = db.QueryRow(`
	SELECT c.Faculty
	FROM Certificate c
	JOIN Student s ON c.CertificateID = s.CertificateID
	WHERE s.OrderOfReceive = ?
	LIMIT 1`, firstOrderAfternoon).Scan(&afternoonFaculty)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	//afternoonExists := err == nil
	rows, err := db.Query(`
        SELECT c.Faculty
        FROM Certificate c
        JOIN Student s ON c.CertificateID = s.CertificateID
        ORDER BY s.OrderOfReceive ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facultyMap := make(map[string]string)
	var faculties []entity.FacultySession
	sessionAssigned := false

	for rows.Next() {
		var faculty string
		if err := rows.Scan(&faculty); err != nil {
			return nil, err
		}
		if _, exists := facultyMap[faculty]; !exists {
			session := "เช้า"
			if faculty == afternoonFaculty || sessionAssigned {
				session = "บ่าย"
				sessionAssigned = true
			}
			facultyMap[faculty] = session
			faculties = append(faculties, entity.FacultySession{Faculty: faculty, SessionOfAnnounce: session})
		}
	}

	// //safety check
	// if !afternoonExists && !sessionAssigned {
	// 	for i := range faculties {
	// 		faculties[i].SessionOfAnnounce = "เช้า"
	// 	}
	// }

	return faculties, nil
}

// Transaction included
func (db *MySQLDB) UpdateAnnouncerQuery(tx *sql.Tx, announcerID int, announcerName, announcerScript, sessionOfAnnounce string, firstOrder, lastOrder int, isBreak bool) error {
	query := `UPDATE Announcer SET AnnouncerName = ?, AnnouncerScript = ?, SessionOfAnnounce = ?, FirstOrder = ?, LastOrder = ?, IsBreak = ? WHERE AnnouncerID = ?`
	_, err := tx.Exec(query, announcerName, announcerScript, sessionOfAnnounce, firstOrder, lastOrder, isBreak, announcerID)
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
	query := `SELECT AnnouncerID, AnnouncerName, AnnouncerScript, SessionOfAnnounce, FirstOrder, LastOrder, IsBreak FROM Announcer`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying announcers: %w", err)
	}
	defer rows.Close()

	announcers := make(map[int]entity.Announcer)

	for rows.Next() {
		var a entity.Announcer
		if err := rows.Scan(&a.AnnouncerID, &a.AnnouncerName, &a.AnnouncerScript, &a.Session, &a.Start, &a.End, &a.IsBreak); err != nil {
			return nil, fmt.Errorf("scanning announcer: %w", err)
		}
		announcers[a.AnnouncerID] = a
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating over announcers: %w", err)
	}

	return announcers, nil
}

// Transaction included
func (db *MySQLDB) InsertAnnouncer(tx *sql.Tx, announcerName, announcerScript, sessionOfAnnounce string, firstOrder, lastOrder int, isBreak bool) error {
	var first, last sql.NullInt64
	first.Int64 = int64(firstOrder)
	last.Int64 = int64(lastOrder)
	if first.Int64 == 0 {
		first.Valid = false
	}
	if last.Int64 == 0 {
		last.Valid = false
	}
	query := `INSERT INTO Announcer (AnnouncerName, AnnouncerScript, SessionOfAnnounce, FirstOrder, LastOrder, IsBreak) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, announcerName, announcerScript, sessionOfAnnounce, firstOrder, lastOrder, isBreak)
	if err != nil {
		return fmt.Errorf("failed to insert announcer: %w", err)
	}
	return nil
}

// Transaction included
func (db *MySQLDB) DeleteAnnouncer(tx *sql.Tx, announcerID int) error {
	query := `DELETE FROM Announcer WHERE AnnouncerID = ?`
	_, err := tx.Exec(query, announcerID)
	if err != nil {
		return fmt.Errorf("failed to delete announcer: %w", err)
	}
	return nil
}
