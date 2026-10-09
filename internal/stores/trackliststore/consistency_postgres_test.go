package trackliststore

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/stores/datastore"
)

const consistencyMigration = "20261005170000_enforce_track_identity_and_positions.sql"

// Integration tests use only MEMOIR_TEST_DATABASE_URL, never DATABASE_URL or
// .env. Create a dedicated database named memoir_test or memoir_test_<suffix>,
// or with a name ending in _test. Each test owns a fresh schema and never
// migrates, truncates, or drops existing tables.
func consistencyTestPostgres(t *testing.T, migrate bool) (context.Context, *pgxpool.Pool, *Store) {
	t.Helper()
	databaseURL := os.Getenv("MEMOIR_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("PostgreSQL integration test requires MEMOIR_TEST_DATABASE_URL pointing to a dedicated test database")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse MEMOIR_TEST_DATABASE_URL: %v", err)
	}
	name := cfg.ConnConfig.Database
	if !strings.HasSuffix(name, "_test") && !strings.HasPrefix(name, "memoir_test_") {
		t.Fatalf("MEMOIR_TEST_DATABASE_URL must name an explicitly dedicated test database ending in _test or starting with memoir_test_; got %q", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	admin, err := pgx.ConnectConfig(ctx, cfg.ConnConfig.Copy())
	if err != nil {
		t.Fatalf("connect to dedicated test database: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })

	schema := "memoir_database_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quotedSchema := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("drop isolated test schema: %v", err)
		}
	})

	cfg.MaxConns = 16
	cfg.ConnConfig.RuntimeParams["search_path"] = quotedSchema
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	cfg.ConnConfig.RuntimeParams["application_name"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test pool: %v", err)
	}

	_, file, _, _ := runtime.Caller(0)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(file), "../../database/migrations/*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, file := range files {
		if !migrate && filepath.Base(file) == consistencyMigration {
			break
		}
		if err := consistencyApplyMigration(ctx, pool, consistencyMigrationSQL(t, filepath.Base(file), "up")); err != nil {
			t.Fatalf("apply %s in isolated schema: %v", filepath.Base(file), err)
		}
	}
	return ctx, pool, New(datastore.New(pool))
}

func consistencyMigrationSQL(t *testing.T, name, direction string) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	contents, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../database/migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	up, down, ok := strings.Cut(string(contents), "-- migrate:down")
	if !ok {
		t.Fatalf("migration %s has no down section", name)
	}
	if direction == "down" {
		return down
	}
	return strings.TrimPrefix(up, "-- migrate:up")
}

// Match the migration runner's transaction boundary, including multi-statement
// SQL and DO blocks. A failed audit must roll back the entire migration.
func consistencyApplyMigration(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func consistencyExecSQL(t *testing.T, ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		t.Fatal(err)
	}
}

func consistencyRowCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM "+pgx.Identifier{table}.Sanitize()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func consistencyMixParams(name string, tracks ...[]string) *AddTracklistParams {
	return &AddTracklistParams{
		Name:    name,
		Date:    "2026-10-05T12:00:00Z",
		URL:     "https://example.com/mixes/" + name,
		Artwork: "artwork.jpg",
		Tracks:  tracks,
	}
}

func consistencyTrackData(name, artist string) []string {
	return []string{name, artist, "128", "am", "House"}
}

func consistencyAddMix(t *testing.T, ctx context.Context, store *Store, params *AddTracklistParams) *Tracklist {
	t.Helper()
	mix, err := store.AddTracklist(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	return mix
}

// Read only the stored fields relevant to creation. The existing detail query
// also embeds a tsvector field that pgx cannot decode into the generated string
// type in binary mode; that separate loader issue is outside these tests.
func consistencyReadTracks(ctx context.Context, pool *pgxpool.Pool, mixID string) ([]*db.GetTrackRow, error) {
	rows, err := pool.Query(ctx, `SELECT t.id, t.artist, t.name, t.genre, t.bpm, t.key, t.created, t.updated
		FROM tracks AS t JOIN tracklist_tracks AS tt ON tt.track_id = t.id
		WHERE tt.tracklist_id = $1 ORDER BY tt.track_number`, mixID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[db.GetTrackRow])
}
