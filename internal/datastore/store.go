package datastore

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	matchFirstCapRegexp = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCapRegexp   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type Store struct {
	*sqlx.DB
}

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

func (s *Store) Close() error {
	return s.DB.Close()
}
