package trek

import (
	"bufio"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	migrationTimeLayout = "20060102150405"
)

// Migration represents a single migration to apply to or rollback from a
// database.
type Migration struct {
	Name     string
	Path     string
	Version  time.Time
	Contents struct {
		Up   string
		Down string
	}
}

// NewMigration returns an initialised migration read from the given file path.
func NewMigration(path string) (*Migration, error) {
	base := filepath.Base(path)
	unparsedVersion := regexp.MustCompile("^\\d+").FindString(base)

	version, err := time.Parse(migrationTimeLayout, unparsedVersion)
	if err != nil {
		return nil, err
	}

	m := &Migration{
		Name:    base,
		Path:    path,
		Version: version,
	}

	if err = m.readContents(); err != nil {
		return nil, err
	}

	return m, nil
}

// Apply applies the migration to the given database using the given driver
// implementation.
func (m *Migration) Apply(driver Driver, db *sql.DB) error {
	tx, _ := db.Begin()

	if _, err := tx.Exec(m.Contents.Up); err != nil {
		tx.Rollback()
		return err
	}

	if err := driver.MarkVersionAsExecuted(tx, m.VersionAsString()); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// Rollback rolls back the migration from the given database using the given
// driver implementation.
func (m *Migration) Rollback(driver Driver, db *sql.DB) error {
	tx, _ := db.Begin()

	if _, err := tx.Exec(m.Contents.Down); err != nil {
		tx.Rollback()
		return err
	}

	if err := driver.UnmarkVersionAsExecuted(tx, m.VersionAsString()); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// HasBeenMigrated checks if the migration has been applied to the given
// database.
func (m *Migration) HasBeenMigrated(driver Driver, db *sql.DB) (bool, error) {
	return driver.HasVersionBeenExecuted(db, m.VersionAsString())
}

// VersionAsString returns the migration version as a string.
func (m *Migration) VersionAsString() string {
	return m.Version.Format(migrationTimeLayout)
}

func (m *Migration) readContents() error {
	f, err := os.Open(m.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	up := true
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "-- up") {
			up = true
		} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "-- down") {
			up = false
		} else {
			if up {
				m.Contents.Up += line + "\n"
			} else if !up {
				m.Contents.Down += line + "\n"
			}
		}
	}

	return nil
}
