package trek

import (
	"database/sql"
	"fmt"
)

// Apply applies the migrations in the given path, to the database using the
// given driver.
func Apply(driverName string, dsn string, migrationsPath string) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := driver.CreateVersionsTable(db); err != nil {
		return err
	}

	migrations, err := LoadMigrationsFromPath(migrationsPath)
	if err != nil {
		return err
	}

	return migrations.Apply(driver, db)
}

// Rollback rolls back the migrations in the given path, to the database using
// the given driver.
func Rollback(driverName string, dsn string, migrationsPath string) error {
	driver, err := getDriver(driverName)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := driver.CreateVersionsTable(db); err != nil {
		return err
	}

	migrations, err := LoadMigrationsFromPath(migrationsPath)
	if err != nil {
		return err
	}

	return migrations.Rollback(driver, db)
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
