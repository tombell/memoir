package datastore

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	matchFirstCapRegexp = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCapRegexp   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type Store struct {
	sqlx.Queryer

	DB *sqlx.DB

	logger *log.Logger
}

func New(dsn string, logger *log.Logger) (*Store, error) {
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

	return &Store{
		Queryer: &loggingQueryer{logger, db},
		DB:      db,
		logger:  logger,
	}, nil
}

func (s *Store) Begin() (*Tx, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	s.logger.Debug("db", "tx", "begin")

	return &Tx{Tx: tx, logger: s.logger}, nil
}
