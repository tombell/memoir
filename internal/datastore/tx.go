package datastore

import (
	"database/sql"

	"github.com/charmbracelet/log"
)

type Tx struct {
	*sql.Tx

	logger log.Logger
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	tx.logger.Debug("db", formatQueryArgs(query, args)...)
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
