package commands

import (
	"database/sql"

	"github.com/spf13/cobra"
	"github.com/tombell/migrate/internal/drivers"
	"github.com/tombell/migrate/internal/migrations"
)

func init() {
	setupSharedFlags(applyCmd)

	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply migrations to the dataabase",
	Long:  "Apply outstanding migrations to the database, skipping those already applied.",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := validateFlags(); err != nil {
			return err
		}

		driver, err := drivers.NewDriver(db)
		if err != nil {
			return err
		}

		db, err := sql.Open(driver.Name(), dsn)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			return err
		}

		if err := driver.CreateSchemaMigrationsTable(db); err != nil {
			return err
		}

		m, err := migrations.NewMigrations(driver, db, migrationsPath)
		if err != nil {
			return err
		}

		return m.Migrate()
	},
}
