package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

func associationIDs(t *testing.T, database *datastore.Store, mixID string) []string {
	t.Helper()
	rows, err := database.Query(context.Background(), `SELECT id::text FROM tracklist_tracks WHERE tracklist_id=$1 ORDER BY track_number`, mixID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids
}

func TestPatchTrackReplacement(t *testing.T) {
	database := testDatabase(t)
	ctx := context.Background()
	store := trackliststore.New(database)
	mix, other := seedMix(t, database, 1), seedMix(t, database, 2)
	sharedID, removedID := "40000000-0000-0000-0000-000000000001", "40000000-0000-0000-0000-000000000002"
	for i, id := range []string{sharedID, removedID} {
		name := "Shared"
		if i == 1 {
			name = "Removed"
		}
		if err := database.AddTrack(ctx, db.AddTrackParams{ID: id, Name: name, Artist: "Artist", Genre: "House", BPM: 120, Key: "8A", Created: time.Now(), Updated: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	for i, association := range []struct {
		mix, track string
		position   int32
	}{{mix.ID, sharedID, 1}, {mix.ID, removedID, 2}, {other.ID, sharedID, 1}} {
		if err := database.AddTracklistTrack(ctx, db.AddTracklistTrackParams{ID: fmt.Sprintf("50000000-0000-0000-0000-%012d", i+1), TracklistID: association.mix, TrackID: association.track, TrackNumber: association.position}); err != nil {
			t.Fatal(err)
		}
	}
	sharedBefore, err := database.GetTrack(ctx, sharedID)
	if err != nil {
		t.Fatal(err)
	}
	otherAssociations := associationIDs(t, database, other.ID)
	updated, err := store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{"name":"Replaced","tracks":[["New","Artist","121.5","9a","House"],["Shared","Artist","140","1A","Trance"],["New","Artist","125","2A","Disco"]]}`))
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Replaced" || updated.TrackCount != 3 || len(updated.Tracks) != 3 || updated.Tracks[1].ID != sharedID || updated.Tracks[0].ID != updated.Tracks[2].ID || updated.Tracks[0].BPM != 121.5 || updated.Tracks[0].Key != "9A" {
		t.Fatalf("replacement: %+v", updated)
	}
	sharedAfter, err := database.GetTrack(ctx, sharedID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sharedBefore, sharedAfter) {
		t.Fatal("shared track metadata changed")
	}
	if _, err := database.GetTrack(ctx, removedID); err != nil {
		t.Fatalf("removed track record was deleted: %v", err)
	}
	if !reflect.DeepEqual(otherAssociations, associationIDs(t, database, other.ID)) {
		t.Fatal("other mix associations changed")
	}
	got, err := store.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(updated, got) {
		t.Fatal("PATCH response and GET differ")
	}
	beforeIDs := associationIDs(t, database, mix.ID)
	if _, err := store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{"artwork":"new.jpg"}`)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeIDs, associationIDs(t, database, mix.ID)) {
		t.Fatal("omitted tracks changed associations")
	}
	before, err := database.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"name":"Invalid","tracks":[]}`, `{"name":"Invalid","tracks":null}`,
		`{"name":"Invalid","tracks":[["Would be new","Artist","120","8A","House"],["short"]]}`,
	} {
		_, err := store.UpdateTracklist(ctx, mix.ID, updateParams(t, body))
		var e *errors.Error
		if !errors.As(err, &e) || e.Status() != http.StatusUnprocessableEntity {
			t.Fatalf("invalid replacement: %v", err)
		}
		after, err := database.GetTracklist(ctx, mix.ID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(beforeIDs, associationIDs(t, database, mix.ID)) {
			t.Fatal("invalid replacement persisted changes")
		}
	}
	// A database failure after the first inserted track must roll back metadata,
	// new track rows, association deletion, and inserted associations together.
	_, err = database.Exec(ctx, fmt.Sprintf(`
CREATE FUNCTION reject_replacement() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.tracklist_id = '%s'::uuid AND NEW.track_number = 2 THEN
    RAISE EXCEPTION 'forced association failure';
  END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER reject_replacement BEFORE INSERT ON tracklist_tracks
FOR EACH ROW EXECUTE FUNCTION reject_replacement();`, mix.ID))
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpdateTracklist(ctx, mix.ID, updateParams(t, `{"name":"Rolled back","tracks":[["Rollback new","Artist","120","8A","House"],["Shared","Artist","120","8A","House"]]}`))
	if err == nil {
		t.Fatal("expected forced failure")
	}
	after, err := database.GetTracklist(ctx, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(beforeIDs, associationIDs(t, database, mix.ID)) {
		t.Fatal("database failure left partial changes")
	}
	var count int64
	if err := database.QueryRow(ctx, `SELECT count(*) FROM tracks WHERE name IN ('Rollback new', 'Would be new')`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d new track records", count)
	}
}

func TestConcurrentPatchIdentityReuse(t *testing.T) {
	database := testDatabase(t)
	store := trackliststore.New(database)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mixes := []*db.Tracklist{seedMix(t, database, 1), seedMix(t, database, 2)}
	start := make(chan struct{})
	results := make([]*trackliststore.Tracklist, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i, mix := range mixes {
		params := updateParams(t, `{"tracks":[["Concurrent","Artist","120","8A","House"]]}`)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results[i], errs[i] = store.UpdateTracklist(ctx, mix.ID, params)
		}()
	}
	close(start)
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(results[0].Tracks) != 1 || len(results[1].Tracks) != 1 || results[0].Tracks[0].ID != results[1].Tracks[0].ID {
		t.Fatalf("concurrent PATCH created duplicate identity: %+v", results)
	}
	var count int64
	if err := database.QueryRow(ctx, `SELECT count(*) FROM tracks`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("want 1 shared track, got %d", count)
	}
}
