package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/datastore"
)

type Mailer interface {
	Send(context.Context, string, string, string) error
}

type Service struct {
	store     *datastore.Store
	config    config.AuthConfig
	mailer    Mailer
	logger    *slog.Logger
	dummyHash string
	work      chan struct{}
	attempts  *Limiter
	emails    chan emailJob
}

type emailJob struct{ userID, purpose, to, subject, body string }

type Session struct {
	Token     string
	User      User
	ExpiresAt time.Time
}

func New(store *datastore.Store, cfg config.AuthConfig, mailer Mailer, logger *slog.Logger) (*Service, error) {
	dummy, err := HashPassword("a dummy password for absent users")
	if err != nil {
		return nil, err
	}
	return &Service{store: store, config: cfg, mailer: mailer, logger: logger, dummyHash: dummy, work: make(chan struct{}, 4), attempts: NewLimiter(10, 15*time.Minute), emails: make(chan emailJob, 64)}, nil
}

func NormalizeEmail(input string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(input))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 || !strings.Contains(email, "@") {
		return "", fmt.Errorf("invalid email address")
	}
	for _, r := range email {
		if r > 127 || r < 33 {
			return "", fmt.Errorf("invalid email address")
		}
	}
	return email, nil
}

func (s *Service) acquire() (func(), error) {
	select {
	case s.work <- struct{}{}:
		return func() { <-s.work }, nil
	default:
		return nil, rateLimitError()
	}
}

func rateLimitError() error {
	return errors.E("auth", http.StatusTooManyRequests, errors.M{"message": {"too many attempts; try again later"}})
}

func (s *Service) Register(ctx context.Context, email, password, name string) error {
	email, err := NormalizeEmail(email)
	if err != nil {
		return errors.E("auth[register]", http.StatusUnprocessableEntity, errors.M{"email": {"must be a valid email address"}})
	}
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || name == "" || utf8.RuneCountInString(name) > 100 {
		return errors.E("auth[register]", http.StatusUnprocessableEntity, errors.M{"display_name": {"must contain 1 to 100 characters"}})
	}
	if !s.attempts.Allow("register:" + email) {
		return rateLimitError()
	}
	release, err := s.acquire()
	if err != nil {
		return err
	}
	defer release()
	hash, err := HashPassword(password)
	if err != nil {
		return errors.E("auth[register]", http.StatusUnprocessableEntity, errors.M{"password": {"must contain 15 to 128 characters"}})
	}
	user, err := s.store.CreateUser(ctx, database.CreateUserParams{ID: uuid.NewString(), Email: email, PasswordHash: hash, DisplayName: name})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil
		}
		return errors.E("auth[register]", err)
	}
	s.sendToken(ctx, user.ID, "verify_email")
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil, errors.E("auth[login]", http.StatusUnauthorized)
	}
	if !s.attempts.Allow("login:" + email) {
		return nil, rateLimitError()
	}
	release, err := s.acquire()
	if err != nil {
		return nil, err
	}
	defer release()
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.E("auth[login]", err)
	}
	found := err == nil
	hash := s.dummyHash
	if found {
		hash = user.PasswordHash
	}
	if !VerifyPassword(hash, password) || !found {
		return nil, errors.E("auth[login]", http.StatusUnauthorized)
	}
	token, err := RandomToken()
	if err != nil {
		return nil, errors.E("auth[login]", err)
	}
	tx, err := s.store.Begin(ctx)
	if err != nil {
		return nil, errors.E("auth[login]", err)
	}
	defer tx.Rollback(ctx)
	q := s.store.WithTx(tx)
	// Serialize session creation with password resets. A hash checked before a
	// reset must never create a session after the reset has revoked sessions.
	current, err := q.LockUser(ctx, user.ID)
	if err != nil {
		return nil, errors.E("auth[login]", err)
	}
	if current.PasswordHash != hash {
		return nil, errors.E("auth[login]", http.StatusUnauthorized)
	}
	expires := time.Now().UTC().Add(s.config.SessionTTL)
	if err := q.CreateSession(ctx, database.CreateSessionParams{TokenHash: tokenHash(token), UserID: current.ID, ExpiresAt: expires}); err != nil {
		return nil, errors.E("auth[login]", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, errors.E("auth[login]", err)
	}
	return &Session{Token: token, User: userModel(*current), ExpiresAt: expires}, nil
}

func (s *Service) Lookup(ctx context.Context, token string) (User, error) {
	if !ValidToken(token) {
		return User{}, errors.E("auth[session]", http.StatusUnauthorized)
	}
	row, err := s.store.GetSession(ctx, tokenHash(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, errors.E("auth[session]", http.StatusUnauthorized)
	}
	if err != nil {
		return User{}, errors.E("auth[session]", err)
	}
	return userModel(row.User), nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if !ValidToken(token) {
		return nil
	}
	if err := s.store.DeleteSession(ctx, tokenHash(token)); err != nil {
		return errors.E("auth[logout]", err)
	}
	return nil
}

func (s *Service) RequestEmail(ctx context.Context, email, purpose string) error {
	if purpose != "verify_email" && purpose != "reset_password" {
		return errors.E("auth[email]", http.StatusBadRequest)
	}
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil
	}
	if !s.attempts.Allow(purpose + ":" + email) {
		return rateLimitError()
	}
	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return errors.E("auth[email]", err)
	}
	s.sendToken(ctx, user.ID, purpose)
	return nil
}

// Delivery failures return the same public response as unknown email addresses.
// Resending remains possible, and an undelivered token cannot invalidate an
// earlier link. Consuming a token invalidates all other links of that purpose.
func (s *Service) sendToken(ctx context.Context, userID, purpose string) {
	if err := s.deliverToken(ctx, userID, purpose); err != nil {
		// SMTP servers can put arbitrary content in errors. Keep credentials,
		// recipient addresses, and links out of exported logs.
		s.logger.Error("account email delivery failed", "user_id", userID, "purpose", purpose)
	}
}

func (s *Service) deliverToken(ctx context.Context, userID, purpose string) error {
	token, err := RandomToken()
	if err != nil {
		return err
	}
	tx, err := s.store.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.store.WithTx(tx)
	user, err := q.LockUser(ctx, userID)
	if err != nil {
		return err
	}
	if purpose == "verify_email" && user.EmailVerifiedAt.Valid {
		return nil
	}
	subject, path, ttl := "Verify your Memoir email", "/verify-email", 24*time.Hour
	if purpose == "reset_password" {
		subject, path, ttl = "Reset your Memoir password", "/reset-password", time.Hour
	}
	if err := q.CreateAccountToken(ctx, database.CreateAccountTokenParams{TokenHash: tokenHash(token), UserID: user.ID, Purpose: purpose, ExpiresAt: time.Now().UTC().Add(ttl)}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	link := s.config.Origin + path + "#token=" + url.QueryEscape(token)
	job := emailJob{userID: user.ID, purpose: purpose, to: user.Email, subject: subject, body: subject + ":\n\n" + link + "\n\nIf you did not request this email, you can ignore it."}
	select {
	case s.emails <- job:
		return nil
	default:
		return fmt.Errorf("account email queue full")
	}
}

// RunEmail keeps SMTP latency out of account-existence responses. The bounded
// queue is best effort; a restart or failed delivery requires a resend.
func (s *Service) RunEmail(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.emails:
			if err := s.mailer.Send(ctx, job.to, job.subject, job.body); err != nil {
				s.logger.Error("account email delivery failed", "user_id", job.userID, "purpose", job.purpose)
			}
		}
	}
}

func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	return s.consume(ctx, token, "verify_email", "")
}

func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	if !ValidToken(token) {
		return errors.E("auth[token]", http.StatusBadRequest, errors.M{"token": {"invalid or expired token"}})
	}
	release, err := s.acquire()
	if err != nil {
		return err
	}
	defer release()
	hash, err := HashPassword(password)
	if err != nil {
		return errors.E("auth[reset]", http.StatusUnprocessableEntity, errors.M{"password": {"must contain 15 to 128 characters"}})
	}
	return s.consume(ctx, token, "reset_password", hash)
}

func (s *Service) consume(ctx context.Context, token, purpose, passwordHash string) error {
	op := errors.Op("auth[token]")
	invalid := func() error {
		return errors.E(op, http.StatusBadRequest, errors.M{"token": {"invalid or expired token"}})
	}
	if !ValidToken(token) {
		return invalid()
	}
	userID, err := s.store.GetAccountToken(ctx, database.GetAccountTokenParams{TokenHash: tokenHash(token), Purpose: purpose})
	if errors.Is(err, pgx.ErrNoRows) {
		return invalid()
	}
	if err != nil {
		return errors.E(op, err)
	}
	tx, err := s.store.Begin(ctx)
	if err != nil {
		return errors.E(op, err)
	}
	defer tx.Rollback(ctx)
	q := s.store.WithTx(tx)
	if _, err := q.LockUser(ctx, userID); err != nil {
		return errors.E(op, err)
	}
	if _, err := q.ConsumeAccountToken(ctx, database.ConsumeAccountTokenParams{TokenHash: tokenHash(token), Purpose: purpose}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invalid()
		}
		return errors.E(op, err)
	}
	if purpose == "verify_email" {
		if err := q.VerifyUserEmail(ctx, userID); err != nil {
			return errors.E(op, err)
		}
	} else {
		if err := q.SetUserPassword(ctx, database.SetUserPasswordParams{ID: userID, PasswordHash: passwordHash}); err != nil {
			return errors.E(op, err)
		}
		if err := q.DeleteUserSessions(ctx, userID); err != nil {
			return errors.E(op, err)
		}
	}
	if err := q.DeleteUserAccountTokens(ctx, database.DeleteUserAccountTokensParams{UserID: userID, Purpose: purpose}); err != nil {
		return errors.E(op, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Service) RunCleanup(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		cleanupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := s.store.DeleteExpiredSessions(cleanupCtx); err != nil && ctx.Err() == nil {
			s.logger.Error("expired session cleanup failed")
		}
		if err := s.store.DeleteExpiredAccountTokens(cleanupCtx); err != nil && ctx.Err() == nil {
			s.logger.Error("expired account token cleanup failed")
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
