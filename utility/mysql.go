package utility

import (
	"capstone/server/entity"
	"database/sql"
	"log"

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

func (db *MySQLDB) QueryStudentsToMap() (map[int]entity.Student, error) {
	query := `
    SELECT 
        s.StudentID,
        s.OrderOfReceive, 
        s.Firstname, 
        s.Surname, 
        CONCAT(c.Degree, ' in ', c.Major, ', ', c.Faculty, ', with honor ', 
            CASE c.Honor 
                WHEN 0 THEN 'none' 
                WHEN 1 THEN 'first honor' 
                WHEN 2 THEN 'second honor' 
            END) AS Certificate, 
        s.NamePronunciation,
        c.Degree,
        c.Faculty,
        c.Major,
        CASE c.Honor 
            WHEN 0 THEN 'none' 
            WHEN 1 THEN 'first honor' 
            WHEN 2 THEN 'second honor' 
        END AS Honor
    FROM 
        Student s
    JOIN 
        Certificate c ON s.CertificateID = c.CertificateID
    ORDER BY 
        s.OrderOfReceive ASC;
    `
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make(map[int]entity.Student)
	counter := 0
	for rows.Next() {
		var s entity.Student
		if err := rows.Scan(
			&s.StudentID,
			&s.OrderOfReceive,
			&s.FirstName,
			&s.LastName,
			&s.Certificate,
			&s.Notes,
			&s.Degree,
			&s.Faculty,
			&s.Major,
			&s.Honor); err != nil {
			return nil, err
		}
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
