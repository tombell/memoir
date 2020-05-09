package trek

import (
	"database/sql"
	"fmt"
)

// Migrate applies the migrations in the given path, to the database using the
// given driver.
func Migrate(driverName, dsn, migrationsPath string) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	if err := driver.CreateVersionsTable(db); err != nil {
		return err
	}

	migrations, err := LoadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	return migrations.Migrate(driver, db)
}

// Rollback rolls back the migrations in the given path, to the database using
// the given driver. If steps is provided as a positive integer, it only rolls
// back that many migrations.
func Rollback(driverName, dsn, migrationsPath string, steps int) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	if err := driver.CreateVersionsTable(db); err != nil {
		return err
	}

	migrations, err := LoadMigrations(migrationsPath)
	if err != nil {
		return err
	}

	return migrations.Rollback(driver, db, steps)
}

func getDriver(driver string) (Driver, error) {
	switch driver {
	case "postgres":
		return &PostgresDriver{}, nil
	case "sqlite3":
		return &SQLiteDriver{}, nil
	default:
		return nil, fmt.Errorf("%v is not a supported driver", driver)
	}
}
