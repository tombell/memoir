package database

import (
	"database/sql"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/pkg/errors"
)

// Database contains a connection to the database.
type Database struct {
	conn *sql.DB
}

// Close closes an open connection to the database.
func (db *Database) Close() error {
	if db.conn != nil {
		if err := db.conn.Close(); err != nil {
			return errors.Wrap(err, "db close failed")
		}
	}

	return nil
}

// Begin begins a new transaction using the database connection.
func (db *Database) Begin() (*sql.Tx, error) {
	err := db.conn.Begin()
	return errors.Wrap(err, "db begin failed")
}

// Open opens a new connection to the database. Pings the database to check the
// connection is good.
func Open(dsn string) (*Database, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "db open failed")
	}

	if err := conn.Ping(); err != nil {
		return nil, errors.Wrap(err, "db ping failed")
	}

	return &Database{conn: conn}, nil
}
