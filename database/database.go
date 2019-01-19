package database

import "database/sql"

// Database ...
type Database struct {
	conn *sql.DB
}

// Close ...
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}

	return nil
}

// Begin ...
func (db *Database) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}

// Open ...
func Open(dsn string) (*Database, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &Database{conn: conn}, nil
}
