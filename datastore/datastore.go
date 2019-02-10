package datastore

import (
	"regexp"
	"strings"

	// Import the Postgres driver for database/sql.
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// DataStore contains an active database handle.
type DataStore struct {
	*sqlx.DB
}

// New returns a new DataStore with a backing database handle.
func New(dsn string) (*DataStore, error) {
	sqlx.NameMapper = toSnakeCase

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "db open failed")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "db ping failed")
	}

	return &DataStore{DB: db}, nil
}

func toSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}
