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
	query := `SELECT StudentID, OrderOfReceive, Firstname, Surname, Notes FROM Student`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make(map[int]entity.Student)
	for rows.Next() {
		var s entity.Student
		var order int
		if err := rows.Scan(&s.StudentID, &order, &s.FirstName, &s.LastName, &s.Notes); err != nil {
			return nil, err
		}
		students[order] = s
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
