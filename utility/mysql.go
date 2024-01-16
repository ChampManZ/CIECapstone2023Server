package utility

import (
	"database/sql"

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
