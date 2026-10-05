package trackliststore

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/errors"
)

func TestAddOrReuseTrackWaitsForConcurrentInsert(t *testing.T) {
	ctx, pool, _ := consistencyTestPostgres(t, true)
	first, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Rollback(context.Background())
	second, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Rollback(context.Background())
	queries := db.New(pool)
	params := db.AddOrReuseTrackParams{
		ID: uuid.NewString(), Artist: "Artist", Name: "Track", Genre: "House",
		BPM: 128, Key: "AM", Created: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}
	firstID, err := queries.WithTx(first).AddOrReuseTrack(ctx, params)
	if err != nil {
		t.Fatal(err)
	}
	var pid int32
	if err := second.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&pid); err != nil {
		t.Fatal(err)
	}

	params.ID, params.Genre, params.BPM, params.Key = uuid.NewString(), "Techno", 140, "CM"
	params.Created, params.Updated = params.Created.Add(time.Hour), params.Updated.Add(time.Hour)
	type result struct {
		id  string
		err error
	}
	finished := make(chan result, 1)
	go func() {
		id, err := queries.WithTx(second).AddOrReuseTrack(ctx, params)
		finished <- result{id: id, err: err}
	}()
	// Observe an actual PostgreSQL lock wait rather than assuming that starting
	// a goroutine caused its insert to overlap the uncommitted first insert.
	deadline := time.Now().Add(5 * time.Second)
	for {
		var waiting bool
		if err := pool.QueryRow(ctx, "SELECT cardinality(pg_blocking_pids($1)) > 0", pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case result := <-finished:
			t.Fatalf("reuse returned before conflicting insert committed: %+v", result)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("second insert did not reach a PostgreSQL lock wait")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := first.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case result := <-finished:
		if result.err != nil || result.id != firstID {
			t.Fatalf("concurrent reuse: id=%s, err=%v; want %s", result.id, result.err, firstID)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := second.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	track, err := queries.GetTrack(ctx, firstID)
	if err != nil {
		t.Fatal(err)
	}
	if consistencyRowCount(t, ctx, pool, "tracks") != 1 || track.Genre != "House" || track.BPM != 128 || track.Key != "AM" ||
		!track.Created.Equal(params.Created.Add(-time.Hour)) || !track.Updated.Equal(params.Updated.Add(-time.Hour)) {
		t.Fatalf("reuse changed the winning insert's metadata: %+v", track)
	}
}

func TestAddTracklistPreservesMetadataCaseAndRepeatedPositions(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	first := consistencyAddMix(t, ctx, store, consistencyMixParams("first", consistencyTrackData("Track", "Artist")))
	firstTracks, err := consistencyReadTracks(ctx, pool, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	existing := firstTracks[0]
	reused := []string{"Track", "Artist", "140", "cm", "Techno"}
	mix := consistencyAddMix(t, ctx, store, consistencyMixParams("second", reused, consistencyTrackData("track", "Artist"),
		consistencyTrackData("Track", "artist"), reused))
	tracks, err := consistencyReadTracks(ctx, pool, mix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 4 || consistencyRowCount(t, ctx, pool, "tracks") != 3 {
		t.Fatalf("expected four positions and three case-sensitive identities: %+v", tracks)
	}
	if !reflect.DeepEqual(tracks[0], existing) || !reflect.DeepEqual(tracks[3], existing) {
		t.Fatalf("reuse overwrote existing metadata or timestamps: before=%+v, after=%+v", existing, tracks[0])
	}
	if tracks[1].ID == existing.ID || tracks[2].ID == existing.ID || tracks[1].ID == tracks[2].ID {
		t.Fatal("case variants must have distinct identities")
	}
	var positions []int32
	err = pool.QueryRow(ctx, "SELECT array_agg(track_number ORDER BY track_number) FROM tracklist_tracks WHERE tracklist_id = $1", mix.ID).Scan(&positions)
	if err != nil || !slices.Equal(positions, []int32{1, 2, 3, 4}) {
		t.Fatalf("track positions changed: %v, err=%v", positions, err)
	}

	// For a new identity repeated in one request, the first occurrence supplies
	// its metadata, just as it did before atomic identity reuse.
	newMix := consistencyAddMix(t, ctx, store, consistencyMixParams("new-repeat", consistencyTrackData("New", "Artist"),
		[]string{"New", "Artist", "150", "dm", "Other"}))
	newTracks, err := consistencyReadTracks(ctx, pool, newMix.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(newTracks) != 2 || newTracks[0].ID != newTracks[1].ID || newTracks[0].Genre != "House" ||
		newTracks[0].BPM != 128 || newTracks[0].Key != "AM" {
		t.Fatalf("first occurrence metadata was not preserved: %+v", newTracks)
	}
}

func TestAddTracklistsConcurrentlyReuseTracksInDifferentOrders(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	// Overlap identity acquisition so reversed input orders would deadlock
	// without the store's consistent lock ordering.
	consistencyExecSQL(t, ctx, pool, `CREATE FUNCTION delay_track_insert() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN PERFORM pg_sleep(0.05); RETURN NEW; END; $$;
		CREATE TRIGGER delay_track_insert BEFORE INSERT ON tracks
		FOR EACH ROW EXECUTE FUNCTION delay_track_insert();`)
	const workers = 10
	start := make(chan struct{})
	finished := make(chan error, workers)
	for i := range workers {
		go func() {
			tracks := [][]string{consistencyTrackData("A", "Artist"), consistencyTrackData("B", "Artist")}
			if i%2 == 1 {
				slices.Reverse(tracks)
			}
			tracks = append(tracks, tracks[0])
			<-start
			mix, err := store.AddTracklist(ctx, consistencyMixParams(fmt.Sprintf("concurrent-%d", i), tracks...))
			if err == nil {
				var saved []*db.GetTrackRow
				saved, err = consistencyReadTracks(ctx, pool, mix.ID)
				if err == nil && (len(saved) != 3 || saved[0].Name != tracks[0][0] ||
					saved[1].Name != tracks[1][0] || saved[0].ID != saved[2].ID) {
					err = fmt.Errorf("mix %s lost the requested order or repeated occurrence", mix.ID)
				}
			}
			finished <- err
		}()
	}
	close(start)
	for range workers {
		if err := <-finished; err != nil {
			t.Errorf("concurrent creation failed: %v", err)
		}
	}
	if consistencyRowCount(t, ctx, pool, "tracks") != 2 || consistencyRowCount(t, ctx, pool, "tracklists") != workers ||
		consistencyRowCount(t, ctx, pool, "tracklist_tracks") != workers*3 {
		t.Fatal("concurrent mixes must share two identities and retain every occurrence")
	}
}

func TestGetTracklistsByTrackListsEachMixOnceAndMatchesTotals(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	first := consistencyAddMix(t, ctx, store, consistencyMixParams("repeated", consistencyTrackData("Shared", "Artist"),
		consistencyTrackData("Other", "Artist"), consistencyTrackData("Shared", "Artist")))
	second := consistencyAddMix(t, ctx, store, consistencyMixParams("single", consistencyTrackData("Shared", "Artist")))
	consistencyAddMix(t, ctx, store, consistencyMixParams("unrelated", consistencyTrackData("Other", "Artist")))
	tracks, err := consistencyReadTracks(ctx, pool, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	trackID := tracks[0].ID
	expected := map[string]int{first.ID: 3, second.ID: 1}
	rows, total, err := store.GetTracklistsByTrack(ctx, trackID, 1, 10)
	if err != nil || total != 2 || len(rows) != 2 {
		t.Fatalf("expected two distinct mixes, got rows=%d, total=%d, err=%v", len(rows), total, err)
	}
	seen := make(map[string]bool)
	for _, mix := range rows {
		if seen[mix.ID] || mix.TrackCount != expected[mix.ID] {
			t.Fatalf("duplicate mix or incorrect occurrence count: %+v", mix)
		}
		seen[mix.ID] = true
	}
	// Both dates are equal, so the ID breaks ties consistently across pages.
	ids := []string{first.ID, second.ID}
	slices.Sort(ids)
	for page := int64(1); page <= 3; page++ {
		rows, total, err = store.GetTracklistsByTrack(ctx, trackID, page, 1)
		if err != nil || total != 2 {
			t.Fatalf("page %d: total=%d, err=%v", page, total, err)
		}
		if page <= 2 {
			if len(rows) != 1 || rows[0].ID != ids[page-1] || rows[0].TrackCount != expected[rows[0].ID] {
				t.Fatalf("page %d returned the wrong mix: %+v", page, rows)
			}
		} else if len(rows) != 0 {
			t.Fatalf("page after last distinct mix should be empty: %+v", rows)
		}
	}
	rows, total, err = store.GetTracklistsByTrack(ctx, uuid.NewString(), 1, 10)
	if err != nil || total != 0 || len(rows) != 0 {
		t.Fatalf("absent track: rows=%d, total=%d, err=%v", len(rows), total, err)
	}
}

func TestAddTracklistRollsBackTracksAndLinksAfterFailure(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	existing := consistencyAddMix(t, ctx, store, consistencyMixParams("existing", consistencyTrackData("Shared", "Artist")))
	existingTracks, err := consistencyReadTracks(ctx, pool, existing.ID)
	if err != nil {
		t.Fatal(err)
	}
	consistencyExecSQL(t, ctx, pool, `CREATE FUNCTION reject_second_position() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN
		  IF NEW.track_number = 2 THEN RAISE EXCEPTION 'injected position failure'; END IF;
		  RETURN NEW;
		END; $$;
		CREATE TRIGGER reject_second_position BEFORE INSERT ON tracklist_tracks
		FOR EACH ROW EXECUTE FUNCTION reject_second_position();`)
	_, err = store.AddTracklist(ctx, consistencyMixParams("failed", consistencyTrackData("New", "Artist"),
		[]string{"Shared", "Artist", "150", "cm", "Other"}))
	if err == nil {
		t.Fatal("injected link failure must fail creation")
	}
	if consistencyRowCount(t, ctx, pool, "tracks") != 1 || consistencyRowCount(t, ctx, pool, "tracklists") != 1 ||
		consistencyRowCount(t, ctx, pool, "tracklist_tracks") != 1 {
		t.Fatal("failed creation left new tracks, links, or a partial mix")
	}
	after, err := consistencyReadTracks(ctx, pool, existing.ID)
	if err != nil || !reflect.DeepEqual(after, existingTracks) {
		t.Fatalf("failed creation changed existing tracks or their metadata: %+v, err=%v", after, err)
	}
	consistencyExecSQL(t, ctx, pool, "DROP TRIGGER reject_second_position ON tracklist_tracks")
	consistencyAddMix(t, ctx, store, consistencyMixParams("failed", consistencyTrackData("New", "Artist")))
}

func TestDeleteTracklistRollsBackLinksAfterFailure(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	mix := consistencyAddMix(t, ctx, store, consistencyMixParams("delete", consistencyTrackData("Shared", "Artist"), consistencyTrackData("Shared", "Artist")))
	consistencyExecSQL(t, ctx, pool, `CREATE FUNCTION reject_mix_delete() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'injected delete failure'; END; $$;
		CREATE TRIGGER reject_mix_delete BEFORE DELETE ON tracklists
		FOR EACH ROW EXECUTE FUNCTION reject_mix_delete();`)
	if err := store.DeleteTracklist(ctx, mix.ID); err == nil {
		t.Fatal("injected mix deletion failure must fail deletion")
	}
	if consistencyRowCount(t, ctx, pool, "tracklists") != 1 || consistencyRowCount(t, ctx, pool, "tracklist_tracks") != 2 ||
		consistencyRowCount(t, ctx, pool, "tracks") != 1 {
		t.Fatal("failed deletion removed mix, links, or identities")
	}
	consistencyExecSQL(t, ctx, pool, "DROP TRIGGER reject_mix_delete ON tracklists")
	if err := store.DeleteTracklist(ctx, mix.ID); err != nil {
		t.Fatal(err)
	}
	if consistencyRowCount(t, ctx, pool, "tracklists") != 0 || consistencyRowCount(t, ctx, pool, "tracklist_tracks") != 0 ||
		consistencyRowCount(t, ctx, pool, "tracks") != 1 {
		t.Fatal("successful deletion must remove every occurrence and retain the track identity")
	}
}

func TestDeleteTracklistConcurrentlyReturnsOneNotFound(t *testing.T) {
	ctx, pool, store := consistencyTestPostgres(t, true)
	mix := consistencyAddMix(t, ctx, store, consistencyMixParams("delete", consistencyTrackData("Shared", "Artist"), consistencyTrackData("Shared", "Artist")))
	remaining := consistencyAddMix(t, ctx, store, consistencyMixParams("remaining", consistencyTrackData("Shared", "Artist")))
	start := make(chan struct{})
	finished := make(chan error, 2)
	for range 2 {
		go func() {
			<-start
			finished <- store.DeleteTracklist(ctx, mix.ID)
		}()
	}
	close(start)
	successes, missing := 0, 0
	for range 2 {
		err := <-finished
		if err == nil {
			successes++
			continue
		}
		var storeErr *errors.Error
		if errors.As(err, &storeErr) && storeErr.Status() == http.StatusNotFound {
			missing++
		} else {
			t.Errorf("unexpected concurrent delete error: %v", err)
		}
	}
	if successes != 1 || missing != 1 {
		t.Fatalf("expected one deletion and one not found, got success=%d missing=%d", successes, missing)
	}
	if consistencyRowCount(t, ctx, pool, "tracklists") != 1 || consistencyRowCount(t, ctx, pool, "tracklist_tracks") != 1 ||
		consistencyRowCount(t, ctx, pool, "tracks") != 1 {
		t.Fatal("concurrent deletion affected a shared track or another mix")
	}
	if _, err := db.New(pool).GetTracklist(ctx, remaining.ID); err != nil {
		t.Fatalf("remaining mix was affected: %v", err)
	}
}
