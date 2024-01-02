package datastore

import (
	"database/sql"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Tx struct {
	*sql.Tx

	logger *log.Logger
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	oldStyle := log.DefaultStyles()
	newStyle := log.DefaultStyles()
	newStyle.Value = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "208", Dark: "192"})
	tx.logger.SetStyles(newStyle)
	tx.logger.Debug("db", formatQueryArgs(query, args)...)
	tx.logger.SetStyles(oldStyle)
	return tx.Tx.Exec(query, args...)
}

func (tx *Tx) Commit() error {
	tx.logger.Debug("db", "tx", "commit")
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	tx.logger.Debug("db", "tx", "rollback")
	return tx.Tx.Rollback()
}
