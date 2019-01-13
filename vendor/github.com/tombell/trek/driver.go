package trek

import (
	"database/sql"
)

// Driver ...
type Driver interface {
	CreateVersionsTable(db *sql.DB) error
	HasVersionBeenExecuted(db *sql.DB, version string) (bool, error)
	MarkVersionAsExecuted(tx *sql.Tx, version string) error
	UnmarkVersionAsExecuted(tx *sql.Tx, version string) error
}

// PostgresDriver is a PostgreSQL implementation of the Driver interface.
type PostgresDriver struct{}

// SQLiteDriver is a SQLite implementation of the Driver interface.
type SQLiteDriver struct {
	PostgresDriver
}

// CreateVersionsTable creates the table for storing migrated versions if it
// doesn't exist.
func (d *PostgresDriver) CreateVersionsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS database_versions(version TEXT);`)
	return err
}

// HasVersionBeenExecuted checks if the given version has been executed.
func (d *PostgresDriver) HasVersionBeenExecuted(db *sql.DB, version string) (bool, error) {
	var count int

	row := db.QueryRow("SELECT COUNT(*) FROM database_versions WHERE version=$1", version)
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// MarkVersionAsExecuted marks a given version as been executed.
func (d *PostgresDriver) MarkVersionAsExecuted(tx *sql.Tx, version string) error {
	_, err := tx.Exec("INSERT INTO database_versions (version) VALUES ($1)", version)
	return err
}

// UnmarkVersionAsExecuted unmarks a given version as been executed.
func (d *PostgresDriver) UnmarkVersionAsExecuted(tx *sql.Tx, version string) error {
	_, err := tx.Exec("DELETE FROM database_versions WHERE version=$1", version)
	return err
}
