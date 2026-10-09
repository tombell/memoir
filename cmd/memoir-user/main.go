// memoir-user provisions the verified owner for an existing installation.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "account email address")
	name := flag.String("name", "", "public display name")
	claim := flag.Bool("claim-legacy-tracklists", false, "assign all currently unowned tracklists to this account")
	flag.Parse()
	normalized, err := auth.NormalizeEmail(*email)
	if err != nil {
		return err
	}
	displayName := strings.TrimSpace(*name)
	if !utf8.ValidString(displayName) || displayName == "" || utf8.RuneCountInString(displayName) > 100 {
		return fmt.Errorf("name must contain 1 to 100 characters")
	}
	fmt.Fprintln(os.Stderr, "Reading password from stdin. Supply 15 to 128 characters followed by a newline.")
	input, err := io.ReadAll(io.LimitReader(os.Stdin, 514))
	if err != nil {
		return err
	}
	password := strings.TrimSuffix(strings.TrimSuffix(string(input), "\n"), "\r")
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("could not connect to database")
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := database.New(tx)
	user, err := q.CreateUser(ctx, database.CreateUserParams{ID: uuid.NewString(), Email: normalized, PasswordHash: hash, DisplayName: displayName})
	if err != nil {
		return fmt.Errorf("could not create account: %w", err)
	}
	if err := q.VerifyUserEmail(ctx, user.ID); err != nil {
		return err
	}
	var claimed int64
	if *claim {
		tag, err := tx.Exec(ctx, "UPDATE tracklists SET owner_id = $1 WHERE owner_id IS NULL", user.ID)
		if err != nil {
			return err
		}
		claimed = tag.RowsAffected()
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	fmt.Printf("Created verified user %s; assigned %d legacy tracklists.\n", user.ID, claimed)
	return nil
}
