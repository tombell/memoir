package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func TestTracklistRequestErrors(t *testing.T) {
	router := validationRouter()
	const validFields = `"name":"Mix","date":"2026-10-05T12:00:00Z","url":"https://example.com/mix"`
	for _, tc := range []struct {
		name, method, path, contentType, body, field string
		status                                       int
	}{
		{"missing content type", "POST", "/tracklists", "", `{}`, "message", 415},
		{"unsupported content type", "POST", "/tracklists", "text/plain", `{}`, "message", 415},
		{"malformed json", "POST", "/tracklists", "application/json", `{"name":`, "message", 400},
		{"trailing json", "POST", "/tracklists", "application/json", `{} {}`, "message", 400},
		{"invalid json types", "POST", "/tracklists", "application/json", `{"tracks":[[123]]}`, "message", 400},
		{"missing artwork", "POST", "/tracklists", "application/json; charset=utf-8",
			`{` + validFields + `,"tracks":[["Track","Artist","128","Am","House"]]}`, "artwork", 422},
		{"short track row", "POST", "/tracklists", "application/json; charset=utf-8",
			`{` + validFields + `,"artwork":"cover.jpg","tracks":[["Track"]]}`, "tracks[0]", 422},
		{"null track row", "POST", "/tracklists", "application/json",
			`{` + validFields + `,"artwork":"cover.jpg","tracks":[null]}`, "tracks[0]", 422},
		{"invalid bpm", "POST", "/tracklists", "application/json",
			`{` + validFields + `,"artwork":"cover.jpg","tracks":[["Track","Artist","fast","Am","House"]]}`, "tracks[0].bpm", 422},
		{"nonfinite bpm", "POST", "/tracklists", "application/json",
			`{` + validFields + `,"artwork":"cover.jpg","tracks":[["Track","Artist","NaN","Am","House"]]}`, "tracks[0].bpm", 422},
		{"patch unsupported content type", "PATCH", "/tracklists/00000000-0000-0000-0000-000000000000", "text/plain", `{}`, "message", 415},
		{"patch malformed json", "PATCH", "/tracklists/00000000-0000-0000-0000-000000000000", "application/json", `{} garbage`, "message", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", tc.contentType)
			r.Header.Set("API-Token", "test-token")
			assertRequestError(t, router, r, tc.status, tc.field)
		})
	}
}

func TestOversizedJSONRequestErrors(t *testing.T) {
	router := validationRouter()
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		for _, streaming := range []bool{false, true} {
			path := "/tracklists"
			if method == http.MethodPatch {
				path += "/00000000-0000-0000-0000-000000000000"
			}
			body := "{}" + strings.Repeat(" ", int(payload.MaxJSONBodyBytes)-1)
			r := httptest.NewRequest(method, path, strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json; charset=utf-8")
			r.Header.Set("API-Token", "test-token")
			if streaming {
				r.ContentLength = -1
			}
			assertRequestError(t, router, r, http.StatusRequestEntityTooLarge, "message")
		}
	}
}

func TestListEndpointsRejectInvalidQueryValues(t *testing.T) {
	router := validationRouter()
	for _, path := range []string{"/tracklists", "/tracks/search", "/tracks/mostplayed"} {
		for _, tc := range []struct {
			field, value string
		}{
			{"page", ""}, {"page", "0"}, {"page", "-1"}, {"page", "invalid"},
			{"page", "9223372036854775807"}, {"page", "9223372036854775808"},
			{"page", "214748366"},
			{"per_page", ""}, {"per_page", "0"}, {"per_page", "-1"},
			{"per_page", "invalid"}, {"per_page", "101"}, {"per_page", "2147483648"},
		} {
			t.Run(path+"/"+tc.field+"/"+tc.value, func(t *testing.T) {
				query := url.Values{tc.field: {tc.value}}
				r := httptest.NewRequest(http.MethodGet, path+"?"+query.Encode(), nil)
				assertRequestError(t, router, r, http.StatusBadRequest, tc.field)
			})
		}
	}

	r := httptest.NewRequest(http.MethodGet, "/tracklists?track_id=invalid&per_page=101", nil)
	assertRequestError(t, router, r, http.StatusBadRequest, "per_page")
	r = httptest.NewRequest(http.MethodGet, "/tracklists?page=%zz", nil)
	assertRequestError(t, router, r, http.StatusBadRequest, "message")
}

func validationRouter() *http.ServeMux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{}
	cfg.API.Token = "test-token"
	router := http.NewServeMux()
	// Nil datastores ensure every rejected request is handled before database
	// access. No database, AWS configuration, or network calls are needed.
	routes(logger, router, cfg, trackliststore.New(nil), trackstore.New(nil), nil)
	return router
}

func assertRequestError(t *testing.T, router http.Handler, r *http.Request, status int, field string) {
	t.Helper()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("HTTP status = %d, want %d; body = %s", w.Code, status, w.Body.String())
	}
	var response struct {
		Errors map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("invalid error response: %v; body = %s", err, w.Body.String())
	}
	if len(response.Errors[field]) == 0 {
		t.Fatalf("response errors = %v, want field %q", response.Errors, field)
	}
}
