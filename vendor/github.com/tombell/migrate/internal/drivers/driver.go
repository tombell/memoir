package drivers

import (
	"database/sql"
	"fmt"
)

type Driver interface {
	Name() string
	CreateSchemaMigrationsTable(db *sql.DB) error
	HasMigrationBeenApplied(db *sql.DB, version string) (bool, error)
	MarkMigrationAsApplied(tx *sql.Tx, version string) error
	UnmarkMigrationAsApplied(tx *sql.Tx, version string) error
}

func NewDriver(driver string) (Driver, error) {
	switch driver {
	case "postgresql", "postgres", "psql":
		return &PostgresSQL{}, nil
	case "sqlite":
		return &SQLite{}, nil
	}
	return nil, fmt.Errorf("unsupported database driver (%s)", driver)
}
