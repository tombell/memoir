package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/stores/datastore"
)

// Every test uses a fresh schema inside an explicitly supplied test database.
// The name guard prevents accidentally applying fixtures to a user database.
func testDatabase(t *testing.T) *datastore.Store {
	t.Helper()
	dsn := os.Getenv("MEMOIR_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MEMOIR_TEST_DATABASE_URL is unset; PostgreSQL integration test not run")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Database != "memoir_test" && !strings.HasPrefix(cfg.ConnConfig.Database, "memoir_test_") {
		t.Fatal("MEMOIR_TEST_DATABASE_URL must name memoir_test or memoir_test_*; refusing to use a user database")
	}
	ctx := context.Background()
	admin, err := pgx.ConnectConfig(ctx, cfg.ConnConfig)
	if err != nil {
		t.Fatal(err)
	}
	schema := "pagination_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		admin.Close(ctx)
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close(ctx)
	})
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	migrations, err := filepath.Glob("../database/migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, path := range migrations {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		up, _, _ := strings.Cut(string(data), "-- migrate:down")
		if _, err := pool.Exec(ctx, up); err != nil {
			t.Fatalf("apply %s: %v", path, err)
		}
	}
	return datastore.New(pool)
}
