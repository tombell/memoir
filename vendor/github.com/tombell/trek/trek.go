package trek

import (
	"database/sql"
	"fmt"
	"log"
)

// Apply applies the migrations in the given path, to the database using the
// given driver.
func Apply(logger *log.Logger, driverName, dsn, migrationsPath string) error {
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

	return migrations.Apply(logger, driver, db)
}

// Rollback rolls back the migrations in the given path, to the database using
// the given driver.
func Rollback(logger *log.Logger, driverName, dsn, migrationsPath string) error {
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

	return migrations.Rollback(logger, driver, db)
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
