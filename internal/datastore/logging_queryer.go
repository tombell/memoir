package datastore

import (
	"database/sql"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jmoiron/sqlx"
)

type loggingQueryer struct {
	logger  *log.Logger
	queryer sqlx.Queryer
}

func (q *loggingQueryer) Query(query string, args ...any) (*sql.Rows, error) {
	q.logQuery(query, args...)
	return q.queryer.Query(query, args...)
}

func (q *loggingQueryer) Queryx(query string, args ...any) (*sqlx.Rows, error) {
	q.logQuery(query, args...)
	return q.queryer.Queryx(query, args...)
}

func (q *loggingQueryer) QueryRowx(query string, args ...any) *sqlx.Row {
	q.logQuery(query, args...)
	return q.queryer.QueryRowx(query, args...)
}

func (q *loggingQueryer) logQuery(query string, args ...any) {
	oldStyle := log.DefaultStyles()
	newStyle := log.DefaultStyles()
	newStyle.Value = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "208", Dark: "192"})
	q.logger.SetStyles(newStyle)
	q.logger.Debug("db", formatQueryArgs(query, args)...)
	q.logger.SetStyles(oldStyle)
}
