package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/controllers/tracklistscontroller"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func TestTracklistResponsesPostgresHTTP(t *testing.T) {
	t.Parallel()
	data := responsesTestDatabase(t)
	configuration := &config.Config{}
	configuration.API.Token = "test-token"
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("server log:\n%s", logs.String())
		}
	})
	api := New(logger, configuration, trackliststore.New(data), trackstore.New(data), nil)
	deleteWrites := make(chan int, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := &responsesBodyWriter{ResponseWriter: w}
		api.router.ServeHTTP(writer, r)
		if r.Method == http.MethodDelete {
			deleteWrites <- writer.writes
		}
	}))
	defer server.Close()

	first := tracklistscontroller.CreateRequest{
		Name: "First mix", Date: "2026-10-05T12:00:00Z",
		URL: "https://example.com/first", Artwork: "first.jpg",
		Tracks: [][]string{
			{"Song A", "Artist A", "128", "am", "House"},
			{"Song B", "Artist B", "130", "bm", "Techno"},
			{"Song A", "Artist A", "128", "am", "House"},
		},
	}
	createdBody := responsesRequest(t, server, http.MethodPost, "/tracklists", first, http.StatusCreated)
	var created tracklistscontroller.CreateResponse
	if err := json.Unmarshal(createdBody, &created); err != nil {
		t.Fatal(err)
	}
	mix := created.Tracklist
	if mix == nil || mix.ID == "" || mix.TrackCount != 3 || len(mix.Tracks) != 3 {
		t.Fatalf("incorrect create data: %s", createdBody)
	}
	if mix.Tracks[0].ID == "" || mix.Tracks[1].ID == "" ||
		mix.Tracks[0].ID != mix.Tracks[2].ID || mix.Tracks[0].ID == mix.Tracks[1].ID ||
		mix.Tracks[0].Name != "Song A" || mix.Tracks[1].Name != "Song B" ||
		mix.Tracks[0].BPM != 128 || mix.Tracks[0].Key != "AM" {
		t.Fatalf("incorrect stored track data/order: %s", createdBody)
	}
	getBody := responsesRequest(t, server, http.MethodGet, "/tracklists/"+mix.ID, nil, http.StatusOK)
	if !bytes.Equal(createdBody, getBody) {
		t.Fatalf("create = %s, GET = %s", createdBody, getBody)
	}

	second := first
	second.Name = "Second mix"
	second.Tracks = [][]string{{"Song A", "Artist A", "99", "g", "Different input genre"}}
	secondBody := responsesRequest(t, server, http.MethodPost, "/tracklists", second, http.StatusCreated)
	var other tracklistscontroller.CreateResponse
	if err := json.Unmarshal(secondBody, &other); err != nil {
		t.Fatal(err)
	}
	if other.Tracklist == nil || other.Tracklist.TrackCount != 1 || len(other.Tracklist.Tracks) != 1 ||
		!reflect.DeepEqual(other.Tracklist.Tracks[0], mix.Tracks[0]) {
		t.Fatalf("reused track lost stored metadata: %s", secondBody)
	}

	duplicateCreate := responsesRequest(t, server, http.MethodPost, "/tracklists", first, http.StatusUnprocessableEntity)
	duplicateUpdate := responsesRequest(t, server, http.MethodPatch, "/tracklists/"+other.Tracklist.ID,
		tracklistscontroller.UpdateRequest{Name: first.Name, Date: first.Date, URL: first.URL},
		http.StatusUnprocessableEntity)
	for _, body := range [][]byte{duplicateCreate, duplicateUpdate} {
		var failure payload.ErrorResponse
		if err := json.Unmarshal(body, &failure); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(failure.Errors, map[string][]string{"name": {"Must be unique"}}) {
			t.Fatalf("duplicate name response = %s", body)
		}
	}
	unchanged := responsesRequest(t, server, http.MethodGet, "/tracklists/"+other.Tracklist.ID, nil, http.StatusOK)
	if !bytes.Equal(secondBody, unchanged) {
		t.Fatalf("failed update changed mix: before=%s after=%s", secondBody, unchanged)
	}

	updatedBody := responsesRequest(t, server, http.MethodPatch, "/tracklists/"+other.Tracklist.ID,
		tracklistscontroller.UpdateRequest{Name: "Updated mix", Date: second.Date, URL: second.URL},
		http.StatusOK)
	var updated tracklistscontroller.UpdateResponse
	if err := json.Unmarshal(updatedBody, &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Tracklist == nil || updated.Tracklist.Name != "Updated mix" ||
		updated.Tracklist.TrackCount != 1 || !reflect.DeepEqual(updated.Tracklist.Tracks, other.Tracklist.Tracks) {
		t.Fatalf("incorrect update data: %s", updatedBody)
	}
	updatedGet := responsesRequest(t, server, http.MethodGet, "/tracklists/"+other.Tracklist.ID, nil, http.StatusOK)
	if !bytes.Equal(updatedBody, updatedGet) {
		t.Fatalf("update = %s, GET = %s", updatedBody, updatedGet)
	}

	responsesRequest(t, server, http.MethodDelete, "/tracklists/"+mix.ID, nil, http.StatusNoContent)
	if writes := <-deleteWrites; writes != 0 {
		t.Fatalf("204 attempted %d body writes", writes)
	}
	responsesRequest(t, server, http.MethodGet, "/tracklists/"+mix.ID, nil, http.StatusNotFound)
}

func responsesRequest(t *testing.T, server *httptest.Server, method, path string, input any, status int) []byte {
	t.Helper()
	var requestBody io.Reader
	if input != nil {
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		requestBody = bytes.NewReader(body)
	}
	request, err := http.NewRequest(method, server.URL+path, requestBody)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("API-Token", "test-token")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("%s %s returned %d %s, want %d", method, path, response.StatusCode, body, status)
	}
	if status == http.StatusNoContent {
		if len(body) != 0 || response.Header.Get("Content-Type") != "" {
			t.Fatalf("204 body=%q Content-Type=%q", body, response.Header.Get("Content-Type"))
		}
	} else if response.Header.Get("Content-Type") != "application/json" || !json.Valid(body) {
		t.Fatalf("expected JSON: headers=%v body=%q", response.Header, body)
	}
	return body
}

// Each test migrates and drops only its own schema in an explicit test database.
func responsesTestDatabase(t *testing.T) *datastore.Store {
	t.Helper()
	dsn := os.Getenv("MEMOIR_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MEMOIR_TEST_DATABASE_URL is unset; PostgreSQL response integration test not run")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ConnConfig.Database != "memoir_test" && !strings.HasPrefix(cfg.ConnConfig.Database, "memoir_test_") {
		t.Fatal("MEMOIR_TEST_DATABASE_URL must name memoir_test or memoir_test_*")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	admin, err := pgx.ConnectConfig(ctx, cfg.ConnConfig.Copy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	schema := "memoir_responses_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	migrations, err := filepath.Glob("../database/migrations/*.sql")
	if err != nil || len(migrations) == 0 {
		t.Fatalf("find migrations: %v", err)
	}
	for _, migration := range migrations {
		contents, err := os.ReadFile(migration)
		if err != nil {
			t.Fatal(err)
		}
		up, _, _ := strings.Cut(string(contents), "-- migrate:down")
		if _, err := pool.Exec(ctx, up); err != nil {
			t.Fatalf("apply %s in isolated schema: %v", migration, err)
		}
	}
	var version string
	if err := pool.QueryRow(ctx, "SHOW server_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	t.Logf("PostgreSQL %s, isolated schema: %s", version, schema)
	return datastore.New(pool)
}

type responsesBodyWriter struct {
	http.ResponseWriter
	writes int
}

func (w *responsesBodyWriter) Write(body []byte) (int, error) {
	w.writes++
	return w.ResponseWriter.Write(body)
}
