package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

func seedMix(t *testing.T, database *datastore.Store, n int) *db.Tracklist {
	t.Helper()
	id := fmt.Sprintf("30000000-0000-0000-0000-%012d", n)
	mix, err := database.AddTracklist(context.Background(), db.AddTracklistParams{
		ID: id, Name: fmt.Sprintf("Mix %d", n), Date: time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC),
		URL: "https://example.com/mix", Artwork: "cover.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	return mix
}

func updateParams(t *testing.T, body string) *trackliststore.UpdateTracklistParams {
	t.Helper()
	var params trackliststore.UpdateTracklistParams
	if err := json.Unmarshal([]byte(body), &params); err != nil {
		t.Fatal(err)
	}
	return &params
}

func TestPatchMetadata(t *testing.T) {
	database := testDatabase(t)
	ctx := context.Background()
	store := trackliststore.New(database)
	mix := seedMix(t, database, 1)
	trackID := "40000000-0000-0000-0000-000000000001"
	if err := database.AddTrack(ctx, db.AddTrackParams{ID: trackID, Name: "Signal", Artist: "Artist", Genre: "House", BPM: 120, Key: "8A", Created: time.Now(), Updated: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := database.AddTracklistTrack(ctx, db.AddTracklistTrackParams{ID: "50000000-0000-0000-0000-000000000001", TracklistID: mix.ID, TrackID: trackID, TrackNumber: 1}); err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{"name":"Renamed"}`))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed" || updated.Artwork != mix.Artwork || updated.URL != mix.URL || !updated.Date.Equal(mix.Date) || updated.TrackCount != 1 || len(updated.Tracks) != 1 || updated.Tracks[0].ID != trackID {
		t.Fatalf("name-only PATCH: %+v", updated)
	}
	updated, err = store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{"artwork":"replacement.png"}`))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Artwork != "replacement.png" || updated.Name != "Renamed" || updated.URL != mix.URL || !updated.Date.Equal(mix.Date) || updated.Tracks[0].ID != trackID {
		t.Fatalf("artwork-only PATCH: %+v", updated)
	}
	got, err := store.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(updated, got) {
		t.Fatalf("PATCH and GET differ: %+v / %+v", updated, got)
	}
	updated, err = store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{}`))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Renamed" || updated.Artwork != "replacement.png" || updated.TrackCount != 1 {
		t.Fatalf("empty PATCH: %+v", updated)
	}
	before, err := database.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{"name":null}`, `{"artwork":""}`, `{"date":"invalid","name":"Do not persist"}`} {
		_, err := store.UpdateTracklist(ctx, mix.ID, updateParams(t, body))
		var e *errors.Error
		if !errors.As(err, &e) || e.Status() != http.StatusUnprocessableEntity {
			t.Fatalf("invalid PATCH %s: %v", body, err)
		}
		after, err := database.GetTracklist(ctx, mix.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("invalid PATCH persisted changes: %+v / %+v", before, after)
		}
	}
	other := seedMix(t, database, 2)
	_, err = store.UpdateTracklist(ctx, mix.ID, updateParams(t, fmt.Sprintf(`{"name":%q,"artwork":"do-not-persist.jpg"}`, other.Name)))
	var e *errors.Error
	if !errors.As(err, &e) || e.Status() != http.StatusUnprocessableEntity || len(e.Message()["name"]) == 0 {
		t.Fatalf("duplicate name: %v", err)
	}
	after, err := database.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("duplicate name persisted metadata")
	}
	// Legacy mixes without associations must not panic when updated.
	updated, err = store.UpdateTracklist(ctx, other.ID, updateParams(t, `{"artwork":"legacy.png"}`))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Artwork != "legacy.png" || updated.TrackCount != 0 {
		t.Fatalf("empty legacy mix: %+v", updated)
	}
	_, err = store.UpdateTracklist(ctx, "30000000-0000-0000-0000-000000000099", updateParams(t, `{"name":"Missing"}`))
	if !errors.As(err, &e) || e.Status() != http.StatusNotFound {
		t.Fatalf("missing mix: %v", err)
	}
}
