package database

import "database/sql"

// Database ...
type Database struct {
	conn *sql.DB
}

// Close ...
func (d *Database) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}

	return nil
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
