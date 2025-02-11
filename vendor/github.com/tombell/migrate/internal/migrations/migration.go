package migrations

import (
	"bufio"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tombell/migrate/internal/drivers"
)

const (
	migrationTimeLayout = "20060102150405"

	migrationUpComment   = "-- migrate:up"
	migrationDownComment = "-- migrate:down"
)

type Migration struct {
	driver drivers.Driver
	db     *sql.DB

	Name    string
	Path    string
	Version time.Time

	contents struct {
		up   string
		down string
	}
}

func newMigration(driver drivers.Driver, db *sql.DB, path string) (*Migration, error) {
	name := filepath.Base(path)
	unparsedVersion := regexp.MustCompile("^\\d+").FindString(name)

	version, err := time.Parse(migrationTimeLayout, unparsedVersion)
	if err != nil {
		return nil, err
	}

	migration := &Migration{
		driver:  driver,
		db:      db,
		Name:    name,
		Path:    path,
		Version: version,
	}

	if err = migration.readMigrationContents(); err != nil {
		return nil, err
	}

	return migration, nil
}

func (m *Migration) migrate() error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.contents.up); err != nil {
		return err
	}

	if err := m.driver.MarkMigrationAsApplied(tx, m.versionString()); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Migration) rollback() error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(m.contents.down); err != nil {
		return err
	}

	if err := m.driver.UnmarkMigrationAsApplied(tx, m.versionString()); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Migration) hasBeenApplied() (bool, error) {
	return m.driver.HasMigrationBeenApplied(m.db, m.versionString())
}

func (m *Migration) versionString() string {
	return m.Version.Format(migrationTimeLayout)
}

func (m *Migration) readMigrationContents() error {
	f, err := os.Open(m.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	up := true
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), migrationUpComment) {
			up = true
			continue
		}

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), migrationDownComment) {
			up = false
			continue
		}

		if up {
			m.contents.up += line + "\n"
		} else {
			m.contents.down += line + "\n"
		}
	}

	return nil
}
