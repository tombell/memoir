package trackliststore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/tombell/memoir/internal/database"
)

func TestReplacementTracksValidation(t *testing.T) {
	for _, body := range []string{
		`{"tracks":null}`, `{"tracks":[]}`, `{"tracks":{}}`, `{"tracks":[null]}`,
		`{"tracks":[["short"]]}`, `{"tracks":[["a","b","120","8A","House","extra"]]}`,
		`{"tracks":[["","b","120","8A","House"]]}`,
		`{"tracks":[["a","","120","8A","House"]]}`,
		`{"tracks":[["a","b",120,"8A","House"]]}`,
		`{"tracks":[["a","b","120","8A",null]]}`,
		`{"tracks":[["a","b","120","8A","\u0000"]]}`,
	} {
		t.Run(body, func(t *testing.T) {
			params := decodeUpdate(t, body)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid replacement")
			}
		})
	}
	for _, bpm := range []string{"", "bad", "NaN", "Inf", "-Inf", "1e400"} {
		params := decodeUpdate(t, fmt.Sprintf(`{"tracks":[["a","b",%q,"8A","House"]]}`, bpm))
		if err := params.Validate(); err == nil || len(err["tracks[0].bpm"]) == 0 {
			t.Fatalf("invalid BPM %q: %v", bpm, err)
		}
	}
	for _, tt := range []struct {
		column, limit int
		field         string
	}{{0, 256, "name"}, {1, 256, "artist"}, {3, 8, "key"}, {4, 128, "genre"}} {
		row := []string{"a", "b", "120", "8A", "House"}
		row[tt.column] = strings.Repeat("é", tt.limit+1)
		params := UpdateTracklistParams{Tracks: PatchField[TrackRows]{Present: true, Value: TrackRows{row}}}
		if err := params.Validate(); err == nil || len(err["tracks[0]."+tt.field]) == 0 {
			t.Fatalf("invalid %s: %v", tt.field, err)
		}
		row[tt.column] = strings.Repeat("é", tt.limit)
		if err := params.Validate(); err != nil {
			t.Fatalf("valid boundary %s: %v", tt.field, err)
		}
	}
	params := decodeUpdate(t, `{"tracks":[["a","b","0","",""],["c","d","125.5","custom","House"]]}`)
	if err := params.Validate(); err != nil {
		t.Fatal(err)
	}
	if !params.Tracks.Present || len(params.Tracks.Value) != 2 {
		t.Fatalf("missing replacement: %+v", params.Tracks)
	}
}

// The fake executes the generated query calls, recording ordered associations
// and exact identity reuse without requiring PostgreSQL for these regressions.
type replacementDB struct {
	tracks       map[[2]string]db.AddTrackParams
	associations []db.AddTracklistTrackParams
	operations   []string
	failPosition int32
}

func (f *replacementDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	name := strings.Fields(sql)[2]
	f.operations = append(f.operations, name)
	switch name {
	case "LockPatchTrackIdentity":
	case "DeleteTracklistTracks":
		f.associations = nil
	case "AddTrack":
		track := db.AddTrackParams{ID: args[0].(string), Artist: args[1].(string), Name: args[2].(string), Genre: args[3].(string), BPM: args[4].(float64), Key: args[5].(string), Created: args[6].(time.Time), Updated: args[7].(time.Time)}
		identity := [2]string{track.Artist, track.Name}
		if _, exists := f.tracks[identity]; exists {
			return pgconn.CommandTag{}, fmt.Errorf("attempted to replace shared track")
		}
		f.tracks[identity] = track
	case "AddTracklistTrack":
		association := db.AddTracklistTrackParams{ID: args[0].(string), TracklistID: args[1].(string), TrackID: args[2].(string), TrackNumber: args[3].(int32)}
		if association.TrackNumber == f.failPosition {
			return pgconn.CommandTag{}, fmt.Errorf("association failure")
		}
		f.associations = append(f.associations, association)
	default:
		return pgconn.CommandTag{}, fmt.Errorf("unexpected query: %s", sql)
	}
	return pgconn.NewCommandTag("OK"), nil
}

func (f *replacementDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected query")
}

func (f *replacementDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	if !strings.Contains(sql, "-- name: GetTrackByArtistAndName") {
		panic("unexpected query row")
	}
	track, found := f.tracks[[2]string{args[0].(string), args[1].(string)}]
	return replacementRow{track: track, found: found}
}

type replacementRow struct {
	track db.AddTrackParams
	found bool
}

func (r replacementRow) Scan(dest ...any) error {
	if !r.found {
		return pgx.ErrNoRows
	}
	values := []any{r.track.ID, r.track.Artist, r.track.Name, r.track.Genre, r.track.BPM, r.track.Key, r.track.Created, r.track.Updated}
	for i, value := range values {
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(value))
	}
	return nil
}

func TestReplacementReusesSharedTracksAndPreservesOrder(t *testing.T) {
	shared := db.AddTrackParams{ID: "shared", Name: "Shared", Artist: "Artist", BPM: 120, Key: "8A", Genre: "House"}
	fake := &replacementDB{tracks: map[[2]string]db.AddTrackParams{{shared.Artist, shared.Name}: shared}}
	input := TrackRows{
		{"New", "Artist", "121.5", "9a", "House"},
		{"Shared", "Artist", "140", "1A", "Trance"},
		{"New", "Artist", "125", "2A", "Disco"},
	}
	if err := replaceTracklistTracks(context.Background(), db.New(fake), "mix", input); err != nil {
		t.Fatal(err)
	}
	if len(fake.tracks) != 2 || !reflect.DeepEqual(fake.tracks[[2]string{"Artist", "Shared"}], shared) {
		t.Fatalf("shared track changed: %+v", fake.tracks)
	}
	newTrack := fake.tracks[[2]string{"Artist", "New"}]
	if newTrack.Key != "9A" || newTrack.BPM != 121.5 {
		t.Fatalf("new track metadata: %+v", newTrack)
	}
	if len(fake.associations) != 3 {
		t.Fatalf("associations: %+v", fake.associations)
	}
	for i, wantID := range []string{newTrack.ID, shared.ID, newTrack.ID} {
		got := fake.associations[i]
		if got.TrackID != wantID || got.TracklistID != "mix" || got.TrackNumber != int32(i+1) {
			t.Fatalf("association %d: %+v", i, got)
		}
	}
	if fake.operations[0] != "LockPatchTrackIdentity" || fake.operations[1] != "DeleteTracklistTracks" {
		t.Fatalf("mutation order: %v", fake.operations)
	}
}

func TestReplacementPropagatesAssociationFailure(t *testing.T) {
	fake := &replacementDB{tracks: make(map[[2]string]db.AddTrackParams), failPosition: 2}
	err := replaceTracklistTracks(context.Background(), db.New(fake), "mix", TrackRows{{"New", "Artist", "120", "8A", "House"}, {"New", "Artist", "120", "8A", "House"}})
	if err == nil || !strings.Contains(err.Error(), "association failure") {
		t.Fatalf("want transaction failure, got %v", err)
	}
}
