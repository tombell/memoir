package trackliststore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/tombell/memoir/internal/api/payload"
	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
)

func TestAddTracklistResponseMatchesGet(t *testing.T) {
	ctx := context.Background()
	existing := db.Track{
		ID: "d4894ad7-abfc-4d72-9b15-e3143b81c1b4", Name: "Existing", Artist: "Artist",
		Genre: "Stored genre", BPM: 124, Key: "AM",
	}
	tracks := map[string]db.Track{existing.ID: existing}
	var mix db.Tracklist
	var trackIDs []string
	tx := &stubTx{t: t}
	tx.queryRow = func(query string, args []any) pgx.Row {
		switch queryName(query) {
		case "AddTracklist":
			mix = db.Tracklist{
				ID: args[0].(string), Name: args[1].(string), URL: args[2].(string),
				Artwork: args[3].(string), Date: args[4].(time.Time),
			}
			return valueRow{values: tracklistValues(mix)}
		case "GetTrackByArtistAndName":
			for _, track := range tracks {
				if track.Artist == args[0] && track.Name == args[1] {
					return valueRow{values: trackValues(track)}
				}
			}
			return valueRow{err: pgx.ErrNoRows}
		default:
			t.Errorf("unexpected row query: %s", queryName(query))
			return valueRow{err: fmt.Errorf("unexpected query")}
		}
	}
	tx.exec = func(query string, args []any) error {
		switch queryName(query) {
		case "AddTrack":
			track := db.Track{
				ID: args[0].(string), Artist: args[1].(string), Name: args[2].(string),
				Genre: args[3].(string), BPM: args[4].(float64), Key: args[5].(string),
				Created: args[6].(time.Time), Updated: args[7].(time.Time),
			}
			tracks[track.ID] = track
		case "AddTracklistTrack":
			if args[1] != mix.ID || args[3] != int32(len(trackIDs)+1) {
				t.Errorf("incorrect mix or track number: %v", args)
			}
			trackIDs = append(trackIDs, args[2].(string))
		default:
			t.Errorf("unexpected exec query: %s", queryName(query))
		}
		return nil
	}
	tx.query = func(query string, args []any) (pgx.Rows, error) {
		if queryName(query) != "GetTracklistWithTracks" || args[0] != mix.ID {
			t.Errorf("unexpected query: %s %v", queryName(query), args)
		}
		if tx.commits != 1 {
			t.Error("read response before transaction committed")
		}
		var rows [][]any
		for _, id := range trackIDs {
			track := tracks[id]
			values := append(tracklistValues(mix), trackValues(track)...)
			rows = append(rows, values)
		}
		return &valueRows{values: rows}, nil
	}
	store := storeWithTx(tx)
	input := validCreateParams()
	input.Tracks = [][]string{
		{"Existing", "Artist", "90", "a", "Input genre"},
		{"New", "Other", "128", "bm", "House"},
		{"Existing", "Artist", "90", "a", "Input genre"},
	}

	created, err := store.AddTracklist(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	found, err := store.GetTracklist(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	createdJSON, err := json.Marshal(created)
	if err != nil {
		t.Fatal(err)
	}
	foundJSON, err := json.Marshal(found)
	if err != nil {
		t.Fatal(err)
	}
	if string(createdJSON) != string(foundJSON) {
		t.Fatalf("created = %s, GET = %s", createdJSON, foundJSON)
	}
	if created.TrackCount != 3 || len(created.Tracks) != 3 {
		t.Fatalf("count = %d, tracks = %v", created.TrackCount, created.Tracks)
	}
	if created.Tracks[0].ID != existing.ID || created.Tracks[2].ID != existing.ID ||
		created.Tracks[0].BPM != existing.BPM || created.Tracks[0].Genre != existing.Genre {
		t.Fatalf("existing track metadata/order lost: %s", createdJSON)
	}
	if created.Tracks[1].Name != "New" || created.Tracks[1].BPM != 128 || created.Tracks[1].Key != "BM" {
		t.Fatalf("new track metadata lost: %s", createdJSON)
	}
	if tx.commits != 1 || tx.rollbacks != 1 {
		t.Fatalf("commits = %d, deferred rollbacks = %d", tx.commits, tx.rollbacks)
	}
}

func TestTracklistWritesMapDuplicateNameErrors(t *testing.T) {
	for _, operation := range []string{"create", "update"} {
		for _, failure := range []struct {
			name       string
			code       string
			constraint string
			status     int
			messages   errors.M
		}{
			{"duplicate name", "23505", "tracklists_name_key", http.StatusUnprocessableEntity, errors.M{"name": {"Must be unique"}}},
			{"duplicate primary key", "23505", "tracklists_pkey", http.StatusInternalServerError, errors.M{"message": {"something went wrong"}}},
			{"other database error", "23503", "tracklists_name_key", http.StatusInternalServerError, errors.M{"message": {"something went wrong"}}},
		} {
			t.Run(operation+"/"+failure.name, func(t *testing.T) {
				pgErr := &pgconn.PgError{Code: failure.code, ConstraintName: failure.constraint, Message: "private database details"}
				tx := &stubTx{t: t, queryRow: func(query string, _ []any) pgx.Row {
					want := "AddTracklist"
					if operation == "update" {
						want = "UpdateTracklist"
					}
					if queryName(query) != want {
						t.Errorf("query = %s, want %s", queryName(query), want)
					}
					return valueRow{err: fmt.Errorf("database wrapper: %w", pgErr)}
				}}
				store := storeWithTx(tx)
				var err error
				if operation == "create" {
					_, err = store.AddTracklist(context.Background(), validCreateParams())
				} else {
					_, err = store.UpdateTracklist(context.Background(), "4d95cc38-92be-426d-84e9-e1ea6bb5bb91", validUpdateParams())
				}
				if err == nil || !errors.Is(err, pgErr) {
					t.Fatalf("error = %v, want wrapped database error", err)
				}
				recorder := httptest.NewRecorder()
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				payload.WriteError(logger, recorder, err)
				var response payload.ErrorResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if recorder.Code != failure.status || !reflect.DeepEqual(response.Errors, map[string][]string(failure.messages)) {
					t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
				}
				if tx.commits != 0 || tx.rollbacks != 1 {
					t.Fatalf("commits = %d, rollbacks = %d", tx.commits, tx.rollbacks)
				}
			})
		}
	}
}

func TestUpdateTracklistMissingRemainsNotFound(t *testing.T) {
	tx := &stubTx{t: t, queryRow: func(string, []any) pgx.Row {
		return valueRow{err: pgx.ErrNoRows}
	}}
	_, err := storeWithTx(tx).UpdateTracklist(context.Background(), "4d95cc38-92be-426d-84e9-e1ea6bb5bb91", validUpdateParams())
	var reported *errors.Error
	if !errors.As(err, &reported) || reported.Status() != http.StatusNotFound {
		t.Fatalf("error = %v, want 404", err)
	}
}

func validCreateParams() *AddTracklistParams {
	return &AddTracklistParams{
		Name: "Mix", Date: "2026-10-05T00:00:00Z", URL: "https://example.com/mix",
		Artwork: "mix.jpg", Tracks: [][]string{{"Song", "Artist", "128", "AM", "House"}},
	}
}

func validUpdateParams() *UpdateTracklistParams {
	input := validCreateParams()
	return &UpdateTracklistParams{Name: input.Name, Date: input.Date, URL: input.URL}
}

func storeWithTx(tx *stubTx) *Store {
	return &Store{dataStore: &stubDataStore{Queries: db.New(tx), tx: tx}}
}

type stubDataStore struct {
	*db.Queries
	tx *stubTx
}

func (s *stubDataStore) Begin(context.Context) (pgx.Tx, error) {
	return s.tx, nil
}

type stubTx struct {
	pgx.Tx
	t         *testing.T
	queryRow  func(string, []any) pgx.Row
	exec      func(string, []any) error
	query     func(string, []any) (pgx.Rows, error)
	commits   int
	rollbacks int
}

func (tx *stubTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	if tx.queryRow == nil {
		tx.t.Errorf("unexpected row query: %s", queryName(query))
		return valueRow{err: fmt.Errorf("unexpected query")}
	}
	return tx.queryRow(query, args)
}

func (tx *stubTx) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if tx.exec == nil {
		tx.t.Errorf("unexpected exec query: %s", queryName(query))
		return pgconn.CommandTag{}, fmt.Errorf("unexpected query")
	}
	return pgconn.NewCommandTag("INSERT 0 1"), tx.exec(query, args)
}

func (tx *stubTx) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx.query == nil {
		tx.t.Errorf("unexpected query: %s", queryName(query))
		return nil, fmt.Errorf("unexpected query")
	}
	return tx.query(query, args)
}

func (tx *stubTx) Commit(context.Context) error {
	tx.commits++
	return nil
}

func (tx *stubTx) Rollback(context.Context) error {
	tx.rollbacks++
	return nil
}

func queryName(query string) string {
	return strings.Fields(query)[2]
}

type valueRow struct {
	values []any
	err    error
}

func (r valueRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return fmt.Errorf("scan destinations = %d, values = %d", len(dest), len(r.values))
	}
	for i, value := range r.values {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(value))
	}
	return nil
}

type valueRows struct {
	pgx.Rows
	values [][]any
	index  int
}

func (r *valueRows) Next() bool {
	r.index++
	return r.index <= len(r.values)
}

func (r *valueRows) Scan(dest ...any) error {
	return (valueRow{values: r.values[r.index-1]}).Scan(dest...)
}

func (*valueRows) Close() {}

func (*valueRows) Err() error {
	return nil
}

func tracklistValues(mix db.Tracklist) []any {
	return []any{mix.ID, mix.Name, mix.Date, mix.Artwork, mix.URL, mix.Created, mix.Updated}
}

func trackValues(track db.Track) []any {
	return []any{track.ID, track.Artist, track.Name, track.Genre, track.BPM, track.Key, track.Created, track.Updated}
}
