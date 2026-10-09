package middleware

import (
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRF(t *testing.T) {
	cfg := config.AuthConfig{Origin: "https://app.example.test", CookieSecure: true}
	token, err := auth.RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	handler := CSRF(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, tt := range []struct {
		name, method, origin, header, cookie string
		status                               int
	}{
		{"safe read", "GET", "", "", "", 204},
		{"allowed", "POST", cfg.Origin, token, token, 204},
		{"wrong origin", "POST", "https://evil.example.test", token, token, 403},
		{"missing origin", "POST", "", token, token, 403},
		{"missing header", "POST", cfg.Origin, "", token, 403},
		{"missing cookie", "POST", cfg.Origin, token, "", 403},
		{"mismatch", "POST", cfg.Origin, token, "different", 403},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, "/", nil)
			r.Header.Set("Origin", tt.origin)
			r.Header.Set("X-CSRF-Token", tt.header)
			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{Name: cfg.CSRFCookieName(), Value: tt.cookie})
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status %d, want %d", w.Code, tt.status)
			}
		})
	}
}
