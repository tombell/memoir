package database

import (
	"github.com/jmoiron/sqlx"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/pkg/errors"
)

// Database contains a connection to the database.
type Database struct {
	conn *sqlx.DB
}

// Close closes an open connection to the database.
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}

	return nil
}

// Begin begins a new transaction using the database connection.
func (db *Database) Begin() (*sqlx.Tx, error) {
	return db.conn.Beginx()
}

// Open opens a new connection to the database. Pings the database to check the
// connection is good.
func Open(dsn string) (*Database, error) {
	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "db open failed")
	}

	if err := conn.Ping(); err != nil {
		return nil, errors.Wrap(err, "db ping failed")
	}

	return &Database{conn: conn}, nil
}
