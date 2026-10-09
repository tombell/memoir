package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/testdb"
)

type sentMail struct{ to, subject, body string }
type recordingMailer struct {
	messages chan sentMail
	fail     bool
}

func (m *recordingMailer) Send(ctx context.Context, to, subject, body string) error {
	select {
	case m.messages <- sentMail{to, subject, body}:
	case <-ctx.Done():
		return ctx.Err()
	}
	if m.fail {
		return errors.New("delivery failed")
	}
	return nil
}

func newTestService(t *testing.T) (*Service, *recordingMailer) {
	t.Helper()
	pool := testdb.New(t)
	mailer := &recordingMailer{messages: make(chan sentMail, 16)}
	service, err := New(datastore.New(pool), config.AuthConfig{Origin: "https://app.example.test", SessionTTL: time.Hour, CookieSecure: true}, mailer, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go service.RunEmail(ctx)
	return service, mailer
}

func mailToken(t *testing.T, m *recordingMailer) string {
	t.Helper()
	select {
	case message := <-m.messages:
		_, tail, found := strings.Cut(message.body, "#token=")
		if !found {
			t.Fatal("missing token in email")
		}
		return strings.Split(tail, "\n")[0]
	case <-time.After(3 * time.Second):
		t.Fatal("missing account email")
		return ""
	}
}

func errorStatus(err error) int {
	var e interface{ Status() int }
	if errors.As(err, &e) {
		return e.Status()
	}
	return 0
}

func TestAccountLifecycle(t *testing.T) {
	s, mailer := newTestService(t)
	ctx := context.Background()
	password := "a sufficiently long password"
	if err := s.Register(ctx, "  Person@Example.Test ", password, "Person"); err != nil {
		t.Fatal(err)
	}
	verify := mailToken(t, mailer)
	if err := s.Register(ctx, "person@example.test", "a different long password", "Changed"); err != nil {
		t.Fatal(err)
	}
	session, err := s.Login(ctx, "person@example.test", password)
	if err != nil {
		t.Fatal(err)
	}
	if session.User.EmailVerified || session.User.DisplayName != "Person" {
		t.Fatal("duplicate registration changed account")
	}
	if _, err := s.Login(ctx, "person@example.test", "a different long password"); errorStatus(err) != http.StatusUnauthorized {
		t.Fatalf("duplicate replaced password: %v", err)
	}
	var stored []byte
	if err := s.store.QueryRow(ctx, "SELECT token_hash FROM sessions WHERE user_id = $1", session.User.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if string(stored) == session.Token || len(stored) != 32 {
		t.Fatal("raw session token stored")
	}
	if err := s.VerifyEmail(ctx, verify); err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyEmail(ctx, verify); errorStatus(err) != http.StatusBadRequest {
		t.Fatalf("verification reused: %v", err)
	}
	if user, err := s.Lookup(ctx, session.Token); err != nil || !user.EmailVerified {
		t.Fatalf("verification not reflected in session: %+v %v", user, err)
	}
	other, err := s.Login(ctx, "person@example.test", password)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Logout(ctx, session.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(ctx, session.Token); errorStatus(err) != http.StatusUnauthorized {
		t.Fatal("logged-out session still valid")
	}
	if err := s.RequestEmail(ctx, "person@example.test", "reset_password"); err != nil {
		t.Fatal(err)
	}
	reset := mailToken(t, mailer)
	if err := s.VerifyEmail(ctx, reset); errorStatus(err) != http.StatusBadRequest {
		t.Fatal("reset token accepted for verification")
	}
	if err := s.ResetPassword(ctx, reset, "a newly changed password"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(ctx, other.Token); errorStatus(err) != http.StatusUnauthorized {
		t.Fatal("reset did not revoke all sessions")
	}
	if err := s.ResetPassword(ctx, reset, password); errorStatus(err) != http.StatusBadRequest {
		t.Fatal("reset token reused")
	}
	if _, err := s.Login(ctx, "person@example.test", password); errorStatus(err) != http.StatusUnauthorized {
		t.Fatal("old password still valid")
	}
	if _, err := s.Login(ctx, "person@example.test", "a newly changed password"); err != nil {
		t.Fatal(err)
	}
}

func TestExpiredAccountTokenAndSession(t *testing.T) {
	s, mailer := newTestService(t)
	ctx := context.Background()
	if err := s.Register(ctx, "person@example.test", "a sufficiently long password", "Person"); err != nil {
		t.Fatal(err)
	}
	token := mailToken(t, mailer)
	if _, err := s.store.Exec(ctx, "UPDATE account_tokens SET expires_at = NOW() - INTERVAL '1 second'"); err != nil {
		t.Fatal(err)
	}
	if err := s.VerifyEmail(ctx, token); errorStatus(err) != http.StatusBadRequest {
		t.Fatal("expired token accepted")
	}
	session, err := s.Login(ctx, "person@example.test", "a sufficiently long password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.store.Exec(ctx, "UPDATE sessions SET expires_at = NOW() - INTERVAL '1 second'"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(ctx, session.Token); errorStatus(err) != http.StatusUnauthorized {
		t.Fatal("expired session accepted")
	}
}

func TestConcurrentResetConsumesOnlyOneLink(t *testing.T) {
	s, mailer := newTestService(t)
	ctx := context.Background()
	if err := s.Register(ctx, "person@example.test", "a sufficiently long password", "Person"); err != nil {
		t.Fatal(err)
	}
	_ = mailToken(t, mailer)
	var tokens []string
	for range 2 {
		if err := s.RequestEmail(ctx, "person@example.test", "reset_password"); err != nil {
			t.Fatal(err)
		}
		tokens = append(tokens, mailToken(t, mailer))
	}
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, token := range tokens {
		wg.Go(func() { results <- s.ResetPassword(ctx, token, "a newly changed password") })
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if errorStatus(err) != http.StatusBadRequest {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("%d concurrent links succeeded", successes)
	}
}

func TestLoginCannotCreateSessionAfterPasswordChange(t *testing.T) {
	s, mailer := newTestService(t)
	ctx := context.Background()
	password := "a sufficiently long password"
	if err := s.Register(ctx, "person@example.test", password, "Person"); err != nil {
		t.Fatal(err)
	}
	_ = mailToken(t, mailer)
	user, err := s.store.GetUserByEmail(ctx, "person@example.test")
	if err != nil {
		t.Fatal(err)
	}
	newHash, err := HashPassword("a newly changed password")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := s.store.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	q := s.store.WithTx(tx)
	if _, err := q.LockUser(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	var pid int32
	if err := tx.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { _, err := s.Login(ctx, user.Email, password); result <- err }()
	// Wait until login has checked the old password and reached the user lock.
	deadline := time.After(10 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var waiting bool
		if err := s.store.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)))", pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("login bypassed password-change lock: %v", err)
		case <-deadline:
			t.Fatal("login did not reach password-change lock")
		case <-ticker.C:
		}
	}
	if err := q.SetUserPassword(ctx, database.SetUserPasswordParams{ID: user.ID, PasswordHash: newHash}); err != nil {
		t.Fatal(err)
	}
	if err := q.DeleteUserSessions(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if errorStatus(err) != http.StatusUnauthorized {
			t.Fatalf("stale login succeeded: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("login remained blocked")
	}
	var count int
	if err := s.store.QueryRow(ctx, "SELECT count(*) FROM sessions WHERE user_id = $1", user.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("password change left a session created using the old password")
	}
}
