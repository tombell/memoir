package commands

import (
	"database/sql"

	"github.com/spf13/cobra"
	"github.com/tombell/migrate/internal/drivers"
	"github.com/tombell/migrate/internal/migrations"
)

var steps int

func init() {
	setupSharedFlags(rollbackCmd)

	rollbackCmd.Flags().IntVar(&steps, "steps", 1, "number of migrations to rollback")

	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "rollback migrations that have been applied to the database",
	Long:  "Rollback applied migrations, reverting the database to previous states.",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		return m.Rollback(steps)
	},
}
