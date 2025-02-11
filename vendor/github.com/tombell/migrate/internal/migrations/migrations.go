package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/tombell/migrate/internal/drivers"
)

type Migrations []*Migration

func (m Migrations) Len() int           { return len(m) }
func (m Migrations) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m Migrations) Less(i, j int) bool { return m[i].Version.Before(m[j].Version) }

func NewMigrations(driver drivers.Driver, db *sql.DB, migrationsPath string) (Migrations, error) {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("os read dir failed: %w", err)
	}

	var migrations Migrations

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		migration, err := newMigration(driver, db, path.Join(migrationsPath, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("new migration failed: %w", err)
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func (m Migrations) Migrate() error {
	sort.Sort(m)

	for _, migration := range m {
		migrated, err := migration.hasBeenApplied()
		if err != nil {
			return fmt.Errorf("checking if migration has been applied failed: %w", err)
		}

		if migrated {
			continue
		}

		if err := migration.migrate(); err != nil {
			return fmt.Errorf("migration failed to apply: %w", err)
		}
	}

	return nil
}

func (m Migrations) Rollback(steps int) error {
	sort.Sort(sort.Reverse(m))

	if steps <= 0 {
		steps = len(m)
	}

	var applied []*Migration

	for _, migration := range m {
		migrated, err := migration.hasBeenApplied()
		if err != nil {
			return fmt.Errorf("checking if migration has been applied failed: %w", err)
		}

		if migrated {
			applied = append(applied, migration)
		}
	}

	if len(applied) == 0 {
		return nil
	}

	for _, migration := range applied[:steps] {
		if err := migration.rollback(); err != nil {
			return fmt.Errorf("migration failed to rollback: %w", err)
		}
	}

	return nil
}
