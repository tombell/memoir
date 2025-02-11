package commands

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

var (
	migrationsPath string
	db             string
	dsn            string
)

var rootCmd = &cobra.Command{
	Use:     "migrate",
	Version: fmt.Sprintf("%s (%s)", Version, Commit),
	Long:    "Migrate is a database migration tool for PostgreSQL and SQLite",
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

func setupSharedFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&db, "db", "postgresql", "postgresql or sqlite")
	cmd.Flags().StringVar(&dsn, "dsn", "", "data source name")
	cmd.Flags().StringVar(&migrationsPath, "migrations", "./migrations", "path to migrations")
	cmd.MarkFlagRequired("dsn")
	cmd.SilenceErrors = true
}

func validateFlags() error {
	if !slices.Contains([]string{"postgresql", "sqlite"}, db) {
		return fmt.Errorf("db must be 'postgresql' or 'sqlite': %s", db)
	}

	f, err := os.Stat(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("migrations directory does not exist")
		}

		return fmt.Errorf("could not stat the migrations directory: %w", err)
	}

	if !f.IsDir() {
		return errors.New("migrations directory must be a directory")
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
