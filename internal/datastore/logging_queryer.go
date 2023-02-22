package datastore

import (
	"database/sql"

	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
)

type loggingQueryer struct {
	logger  log.Logger
	queryer sqlx.Queryer
}

func (q *loggingQueryer) Query(query string, args ...any) (*sql.Rows, error) {
	q.logger.Debug("db", formatQueryArgs(query, args)...)
	return q.queryer.Query(query, args...)
}

func (q *loggingQueryer) Queryx(query string, args ...any) (*sqlx.Rows, error) {
	q.logger.Debug("db", formatQueryArgs(query, args)...)
	return q.queryer.Queryx(query, args...)
}

func (q *loggingQueryer) QueryRowx(query string, args ...any) *sqlx.Row {
	q.logger.Debug("db", formatQueryArgs(query, args)...)
	return q.queryer.QueryRowx(query, args...)
}
