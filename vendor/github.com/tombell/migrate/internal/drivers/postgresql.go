package drivers

import "database/sql"

type PostgresSQL struct{}

func (p *PostgresSQL) Name() string {
	return "pgx"
}

func (p *PostgresSQL) CreateSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations(version TEXT);")
	return err
}

func (p *PostgresSQL) HasMigrationBeenApplied(db *sql.DB, version string) (bool, error) {
	var count int

	row := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=$1", version)
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (p *PostgresSQL) MarkMigrationAsApplied(tx *sql.Tx, version string) error {
	_, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
	return err
}

func (p *PostgresSQL) UnmarkMigrationAsApplied(tx *sql.Tx, version string) error {
	_, err := tx.Exec("DELETE FROM schema_migrations WHERE version=$1", version)
	return err
}
