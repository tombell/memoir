package database_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/testdb"
)

func TestAccountSchema(t *testing.T) {
	pool := testdb.New(t)
	q := database.New(pool)
	ctx := context.Background()
	user, err := q.CreateUser(ctx, database.CreateUserParams{ID: uuid.NewString(), Email: "user@example.test", PasswordHash: "!", DisplayName: "User"})
	if err != nil {
		t.Fatal(err)
	}
	if user.EmailVerifiedAt != nil {
		t.Fatal("new accounts must be unverified")
	}
	if _, err := q.CreateUser(ctx, database.CreateUserParams{ID: uuid.NewString(), Email: user.Email, PasswordHash: "!", DisplayName: "Duplicate"}); err == nil {
		t.Fatal("duplicate email accepted")
	}
	hash := sha256.Sum256([]byte("session"))
	if err := q.CreateSession(ctx, database.CreateSessionParams{TokenHash: hash[:], UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	session, err := q.GetSession(ctx, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	if session.User.EmailVerifiedAt != nil {
		t.Fatal("session must reflect an unverified account")
	}
	if err := q.VerifyUserEmail(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	user, err = q.GetUser(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if user.EmailVerifiedAt == nil || user.EmailVerifiedAt.IsZero() {
		t.Fatal("verified accounts must have a verification time")
	}
	session, err = q.GetSession(ctx, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	if session.User.EmailVerifiedAt == nil || !session.User.EmailVerifiedAt.Equal(*user.EmailVerifiedAt) {
		t.Fatal("session must reflect the account verification time")
	}
	if _, err := pool.Exec(ctx, "UPDATE sessions SET expires_at = NOW() - INTERVAL '1 second'"); err != nil {
		t.Fatal(err)
	}
	if _, err := q.GetSession(ctx, hash[:]); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expired session: %v", err)
	}
	if err := q.CreateAccountToken(ctx, database.CreateAccountTokenParams{TokenHash: hash[:], UserID: user.ID, Purpose: "verify_email", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	params := database.ConsumeAccountTokenParams{TokenHash: hash[:], Purpose: "verify_email"}
	if got, err := q.ConsumeAccountToken(ctx, params); err != nil || got != user.ID {
		t.Fatalf("consume: %s, %v", got, err)
	}
	if _, err := q.ConsumeAccountToken(ctx, params); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("token reused: %v", err)
	}
}

func TestOwnershipMigrationPreservesLegacyRows(t *testing.T) {
	pool := testdb.Open(t)
	testdb.Apply(t, pool, "", "20191031185712_add_generated_tsvector_column_to_tracks_table.sql")
	ctx := context.Background()
	id := uuid.NewString()
	if _, err := pool.Exec(ctx, "INSERT INTO tracklists VALUES ($1, 'Legacy mix', NOW(), 'art.png', 'https://example.test', NOW(), NOW())", id); err != nil {
		t.Fatal(err)
	}
	testdb.Apply(t, pool, "20191031185712_add_generated_tsvector_column_to_tracks_table.sql", "20261009000100_add_users_and_sessions.sql")
	var name string
	var owner *string
	if err := pool.QueryRow(ctx, "SELECT name, owner_id::text FROM tracklists WHERE id = $1", id).Scan(&name, &owner); err != nil {
		t.Fatal(err)
	}
	if name != "Legacy mix" || owner != nil {
		t.Fatal("legacy data was changed or silently assigned")
	}
}
