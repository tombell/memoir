package datastore

import (
	"fmt"
	"regexp"
	"strings"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

var (
	matchFirstCapRegexp = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCapRegexp   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// Store is a structure for interacting with a database, it contains a handle to
// the underlying database.
type Store struct {
	*sqlx.DB
}

// New returns an initialised Store, that has connected to a database, and
// verified with a ping.
func New(dsn string) (*Store, error) {
	sqlx.NameMapper = func(s string) string {
		snake := matchFirstCapRegexp.ReplaceAllString(s, "${1}_${2}")
		snake = matchAllCapRegexp.ReplaceAllString(snake, "${1}_${2}")
		return strings.ToLower(snake)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return &Store{DB: db}, nil
}

// Close closes the connection to the database.
func (s *Store) Close() error {
	return s.DB.Close()
}

func toSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}
