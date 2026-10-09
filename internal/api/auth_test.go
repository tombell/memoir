package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
	"github.com/tombell/memoir/internal/testdb"
)

type testMailer struct{ bodies chan string }

func (m testMailer) Send(ctx context.Context, _, _, body string) error {
	select {
	case m.bodies <- body:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type apiFixture struct {
	server  *Server
	cfg     *config.Config
	store   *datastore.Store
	emails  chan string
	csrf    string
	cookies []*http.Cookie
}

func newFixture(t *testing.T) *apiFixture {
	t.Helper()
	pool := testdb.New(t)
	store := datastore.New(pool)
	cfg := &config.Config{Auth: config.AuthConfig{Origin: "https://app.example.test", CookieSecure: true, SessionTTL: time.Hour}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	emails := make(chan string, 16)
	accounts, err := auth.New(store, cfg.Auth, testMailer{emails}, logger)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go accounts.RunEmail(ctx)
	f := &apiFixture{server: New(logger, cfg, trackliststore.New(store), trackstore.New(store), artworkstore.New(nil), accounts), cfg: cfg, store: store, emails: emails}
	w := f.request("GET", "/auth/csrf", "")
	if w.Code != 200 {
		t.Fatalf("csrf: %d %s", w.Code, w.Body)
	}
	var result struct {
		Data struct {
			Token string `json:"csrf_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	f.csrf, f.cookies = result.Data.Token, w.Result().Cookies()
	return f
}

func (f *apiFixture) request(method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	r.Header.Set("Origin", f.cfg.Auth.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf)
	for _, cookie := range f.cookies {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	f.server.router.ServeHTTP(w, r)
	return w
}

func (f *apiFixture) emailToken(t *testing.T) string {
	t.Helper()
	select {
	case body := <-f.emails:
		_, tail, _ := strings.Cut(body, "#token=")
		return strings.Split(tail, "\n")[0]
	case <-time.After(3 * time.Second):
		t.Fatal("missing email")
		return ""
	}
}

func (f *apiFixture) registerLogin(t *testing.T, email string) string {
	t.Helper()
	body := `{"email":"` + email + `","password":"a sufficiently long password","display_name":"Person"}`
	if w := f.request("POST", "/auth/register", body); w.Code != 202 {
		t.Fatalf("register: %d %s", w.Code, w.Body)
	}
	token := f.emailToken(t)
	w := f.request("POST", "/auth/login", `{"email":"`+email+`","password":"a sufficiently long password"}`)
	if w.Code != 200 {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	f.cookies = append(f.cookies, w.Result().Cookies()...)
	return token
}

func TestBrowserAccountFlow(t *testing.T) {
	f := newFixture(t)
	verify := f.registerLogin(t, "person@example.test")
	var session *http.Cookie
	for _, cookie := range f.cookies {
		if cookie.Name == f.cfg.Auth.SessionCookieName() {
			session = cookie
		}
	}
	if session == nil || !session.HttpOnly || !session.Secure || session.SameSite != http.SameSiteLaxMode || session.Domain != "" || session.Path != "/" {
		t.Fatal("session cookie lacks required protections")
	}
	if w := f.request("GET", "/auth/me", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"email_verified":false`) || strings.Contains(w.Body.String(), "password") {
		t.Fatalf("me: %d %s", w.Code, w.Body)
	}
	if w := f.request("POST", "/auth/verify-email", `{"token":"`+verify+`"}`); w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("verify: %d %s", w.Code, w.Body)
	}
	if w := f.request("GET", "/auth/me", ""); !strings.Contains(w.Body.String(), `"email_verified":true`) {
		t.Fatal("session did not reflect verified email")
	}
	known := f.request("POST", "/auth/forgot-password", `{"email":"person@example.test"}`)
	unknown := f.request("POST", "/auth/forgot-password", `{"email":"unknown@example.test"}`)
	if known.Code != 202 || unknown.Code != 202 || known.Body.String() != unknown.Body.String() {
		t.Fatal("recovery disclosed account existence")
	}
	reset := f.emailToken(t)
	if w := f.request("POST", "/auth/reset-password", `{"token":"`+reset+`","password":"a newly changed password"}`); w.Code != 204 {
		t.Fatalf("reset: %d %s", w.Code, w.Body)
	}
	if w := f.request("GET", "/auth/me", ""); w.Code != 401 {
		t.Fatal("password reset did not revoke browser session")
	}
	if w := f.request("POST", "/auth/logout", ""); w.Code != 204 || w.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("logout did not clear cookie")
	}
}

func TestAuthRejectsMissingCSRFAndInvalidJSON(t *testing.T) {
	f := newFixture(t)
	for _, body := range []string{"{", `{"email":"person@example.test","unknown":true}`, `{} {}`} {
		if w := f.request("POST", "/auth/register", body); w.Code != 400 {
			t.Fatalf("invalid JSON: %d", w.Code)
		}
	}
	f.csrf = ""
	if w := f.request("POST", "/auth/login", `{}`); w.Code != 403 {
		t.Fatal("login accepted without CSRF header")
	}
}

func TestAuthRateLimitsLogin(t *testing.T) {
	f := newFixture(t)
	for i := range 11 {
		w := f.request("POST", "/auth/login", `{"email":"unknown@example.test","password":"a sufficiently long password"}`)
		want := 401
		if i == 10 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("attempt %d: status %d", i, w.Code)
		}
	}
}
