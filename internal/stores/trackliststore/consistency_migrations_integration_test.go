package trackliststore

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func consistencyInsertLegacyTrack(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name, artist, genre string) string {
	t.Helper()
	id := uuid.NewString()
	consistencyExecSQL(t, ctx, pool, `INSERT INTO tracks (id, name, artist, genre, bpm, key, created, updated)
		VALUES ($1, $2, $3, $4, 128, 'AM', '2020-01-01', '2020-01-02')`, id, name, artist, genre)
	return id
}

func consistencyInsertLegacyMix(t *testing.T, ctx context.Context, pool *pgxpool.Pool, name string) string {
	t.Helper()
	id := uuid.NewString()
	consistencyExecSQL(t, ctx, pool, `INSERT INTO tracklists (id, name, date, url, artwork, created, updated)
		VALUES ($1, $2, '2020-01-01', 'https://example.com', 'artwork.jpg', '2020-01-01', '2020-01-01')`, id, name)
	return id
}

func consistencyInsertLegacyPosition(t *testing.T, ctx context.Context, pool *pgxpool.Pool, mix, track string, position int) string {
	t.Helper()
	id := uuid.NewString()
	consistencyExecSQL(t, ctx, pool, `INSERT INTO tracklist_tracks (id, tracklist_id, track_id, track_number)
		VALUES ($1, $2, $3, $4)`, id, mix, track, position)
	return id
}

func consistencyRequirePGError(t *testing.T, err error, code, constraint string) *pgconn.PgError {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != code {
		t.Fatalf("expected PostgreSQL error %s, got %v", code, err)
	}
	if constraint != "" && pgErr.ConstraintName != constraint {
		t.Fatalf("expected constraint %s, got %s", constraint, pgErr.ConstraintName)
	}
	return pgErr
}

func TestConsistencyMigrationRejectsDuplicateIdentitiesWithoutChangingRows(t *testing.T) {
	ctx, pool, _ := consistencyTestPostgres(t, false)
	first := consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "House")
	second := consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "Techno")
	mix := consistencyInsertLegacyMix(t, ctx, pool, "legacy")
	consistencyInsertLegacyPosition(t, ctx, pool, mix, first, 1)
	consistencyInsertLegacyPosition(t, ctx, pool, mix, second, 2)

	err := consistencyApplyMigration(ctx, pool, consistencyMigrationSQL(t, consistencyMigration, "up"))
	pgErr := consistencyRequirePGError(t, err, "P0001", "")
	if !strings.Contains(pgErr.Message, "duplicate artist/name") ||
		!strings.Contains(pgErr.Detail, first) || !strings.Contains(pgErr.Detail, second) ||
		!strings.Contains(pgErr.Hint, "SELECT artist, name") || !strings.Contains(pgErr.Hint, "tracklist_tracks") {
		t.Fatalf("migration must identify conflicting IDs and a repair query: %+v", pgErr)
	}

	var unchanged int
	err = pool.QueryRow(ctx, `SELECT count(*) FROM tracklist_tracks AS tt
		JOIN tracks AS t ON t.id = tt.track_id
		WHERE tt.tracklist_id = $1 AND
		((t.id = $2 AND t.genre = 'House' AND tt.track_number = 1) OR
		 (t.id = $3 AND t.genre = 'Techno' AND tt.track_number = 2))`, mix, first, second).Scan(&unchanged)
	if err != nil || unchanged != 2 || consistencyRowCount(t, ctx, pool, "tracks") != 2 {
		t.Fatalf("failed migration changed identities or references: count=%d, err=%v", unchanged, err)
	}
	// No constraint from the failed migration may remain installed.
	consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "Other")
	consistencyInsertLegacyPosition(t, ctx, pool, mix, first, 0)
}

func TestConsistencyMigrationAuditsInvalidPositions(t *testing.T) {
	for _, test := range []struct {
		name      string
		positions []int
		message   string
		hint      string
	}{
		{name: "nonpositive", positions: []int{0, -1}, message: "nonpositive positions", hint: "WHERE track_number <= 0"},
		{name: "duplicate", positions: []int{1, 1}, message: "duplicate positions", hint: "GROUP BY tracklist_id, track_number"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, pool, _ := consistencyTestPostgres(t, false)
			track := consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "House")
			mix := consistencyInsertLegacyMix(t, ctx, pool, "legacy")
			for _, position := range test.positions {
				consistencyInsertLegacyPosition(t, ctx, pool, mix, track, position)
			}

			err := consistencyApplyMigration(ctx, pool, consistencyMigrationSQL(t, consistencyMigration, "up"))
			pgErr := consistencyRequirePGError(t, err, "P0001", "")
			if !strings.Contains(pgErr.Message, test.message) ||
				!strings.Contains(pgErr.Detail, mix) || !strings.Contains(pgErr.Hint, test.hint) {
				t.Fatalf("expected actionable position audit, got %+v", pgErr)
			}
			var positions []int32
			err = pool.QueryRow(ctx, "SELECT array_agg(track_number ORDER BY track_number) FROM tracklist_tracks").Scan(&positions)
			if err != nil || len(positions) != 2 || positions[0] != int32(test.positions[1]) || positions[1] != int32(test.positions[0]) {
				t.Fatalf("audit changed positions: %v, err=%v", positions, err)
			}
			// The identity constraint must also be absent after this failure.
			consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "Other")
		})
	}
}

func TestConsistencyMigrationEnforcesIdentityAndPositionConstraints(t *testing.T) {
	ctx, pool, _ := consistencyTestPostgres(t, false)
	track := consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "House")
	// Exact identity stays case-sensitive in both the audit and constraint.
	caseTrack := consistencyInsertLegacyTrack(t, ctx, pool, "Track", "artist", "House")
	consistencyInsertLegacyTrack(t, ctx, pool, "track", "Artist", "House")
	mix := consistencyInsertLegacyMix(t, ctx, pool, "legacy")
	otherMix := consistencyInsertLegacyMix(t, ctx, pool, "other")
	consistencyInsertLegacyPosition(t, ctx, pool, mix, track, 1)
	consistencyInsertLegacyPosition(t, ctx, pool, mix, track, 2)
	consistencyInsertLegacyPosition(t, ctx, pool, otherMix, track, 1)
	if err := consistencyApplyMigration(ctx, pool, consistencyMigrationSQL(t, consistencyMigration, "up")); err != nil {
		t.Fatal(err)
	}

	_, err := pool.Exec(ctx, `INSERT INTO tracks (id, name, artist, genre, bpm, key, created, updated)
		VALUES ($1, 'Track', 'Artist', 'Different', 120, 'CM', NOW(), NOW())`, uuid.NewString())
	consistencyRequirePGError(t, err, "23505", "tracks_artist_name_key")
	for _, position := range []int{0, -1} {
		_, err = pool.Exec(ctx, `INSERT INTO tracklist_tracks (id, tracklist_id, track_id, track_number)
			VALUES ($1, $2, $3, $4)`, uuid.NewString(), mix, track, position)
		consistencyRequirePGError(t, err, "23514", "tracklist_tracks_positive_position")
	}
	for _, trackID := range []string{track, caseTrack} {
		_, err = pool.Exec(ctx, `INSERT INTO tracklist_tracks (id, tracklist_id, track_id, track_number)
			VALUES ($1, $2, $3, 1)`, uuid.NewString(), mix, trackID)
		consistencyRequirePGError(t, err, "23505", "tracklist_tracks_tracklist_position_key")
	}
	consistencyInsertLegacyPosition(t, ctx, pool, mix, track, 3)
	if consistencyRowCount(t, ctx, pool, "tracks") != 3 || consistencyRowCount(t, ctx, pool, "tracklist_tracks") != 4 {
		t.Fatal("migration must preserve case variants and repeated track occurrences")
	}

	if err := consistencyApplyMigration(ctx, pool, consistencyMigrationSQL(t, consistencyMigration, "down")); err != nil {
		t.Fatal(err)
	}
	consistencyInsertLegacyTrack(t, ctx, pool, "Track", "Artist", "Different")
	consistencyInsertLegacyPosition(t, ctx, pool, mix, track, 0)
	consistencyInsertLegacyPosition(t, ctx, pool, mix, track, 1)
}
