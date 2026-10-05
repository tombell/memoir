package integration_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tombell/memoir/internal/controllers"
	"github.com/tombell/memoir/internal/controllers/searchcontroller"
	"github.com/tombell/memoir/internal/controllers/trackscontroller"
	db "github.com/tombell/memoir/internal/database"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func TestTrackPagination(t *testing.T) {
	database := testDatabase(t)
	ctx := context.Background()
	store := trackstore.New(database)
	ids := []string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000003",
	}
	// Insert in reverse ID order so tied results must use the SQL tie-breaker.
	for i := len(ids) - 1; i >= 0; i-- {
		if err := database.AddTrack(ctx, db.AddTrackParams{
			ID: ids[i], Name: "Signal", Artist: fmt.Sprintf("Artist %d", i),
			Genre: "House", BPM: 120, Key: "8A", Created: time.Now(), Updated: time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
	}
	mixes := []string{"10000000-0000-0000-0000-000000000001", "10000000-0000-0000-0000-000000000002"}
	for i, id := range mixes {
		if _, err := database.AddTracklist(ctx, db.AddTracklistParams{ID: id, Name: fmt.Sprintf("Mix %d", i), URL: "https://example.com/mix", Artwork: "cover.jpg", Date: time.Now()}); err != nil {
			t.Fatal(err)
		}
	}
	// Track 1 appears three times in one mix; track 2 appears in two mixes.
	for i, association := range []struct{ mix, track int }{{0, 0}, {0, 0}, {0, 0}, {0, 1}, {1, 1}} {
		if err := database.AddTracklistTrack(ctx, db.AddTracklistTrackParams{
			ID: fmt.Sprintf("20000000-0000-0000-0000-%012d", i+1), TracklistID: mixes[association.mix], TrackID: ids[association.track], TrackNumber: int32(i + 1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	search := searchcontroller.Tracks(store)
	for i, id := range ids {
		response, err := search(ctx, searchcontroller.SearchTracksRequest{Query: "signal", Page: fmt.Sprint(i + 1), PerPage: "1"})
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Tracks) != 1 || response.Tracks[0].ID != id || response.Meta != controllers.NewMeta(3, int64(i+1), 1) {
			t.Fatalf("search page %d: %+v, tracks %+v", i+1, response.Meta, response.Tracks)
		}
	}
	for _, query := range []string{"signal", "missing", ""} {
		response, err := search(ctx, searchcontroller.SearchTracksRequest{Query: query, Page: "5", PerPage: "1"})
		if err != nil {
			t.Fatal(err)
		}
		wantTotal := int64(0)
		if query == "signal" {
			wantTotal = 3
		}
		if response.Tracks == nil || len(response.Tracks) != 0 || response.Meta != controllers.NewMeta(wantTotal, 5, 1) {
			t.Fatalf("empty search %q: %+v", query, response)
		}
	}
	mostPlayed := trackscontroller.MostPlayed(store)
	for i, want := range []struct {
		id     string
		played int64
	}{{ids[1], 2}, {ids[0], 1}} {
		response, err := mostPlayed(ctx, trackscontroller.MostPlayedRequest{Page: fmt.Sprint(i + 1), PerPage: "1"})
		if err != nil {
			t.Fatal(err)
		}
		if len(response.Tracks) != 1 || response.Tracks[0].ID != want.id || response.Tracks[0].Played != want.played || response.Meta != controllers.NewMeta(2, int64(i+1), 1) {
			t.Fatalf("most played page %d: %+v", i+1, response)
		}
	}
	response, err := mostPlayed(ctx, trackscontroller.MostPlayedRequest{Page: "3", PerPage: "1"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(response.Tracks, []*trackstore.Track{}) || response.Meta.Total != 2 {
		t.Fatalf("past final page: %+v", response)
	}
	// Give track 3 one appearance, tying track 1, and verify the secondary order.
	if err := database.AddTracklistTrack(ctx, db.AddTracklistTrackParams{ID: "20000000-0000-0000-0000-000000000006", TracklistID: mixes[1], TrackID: ids[2], TrackNumber: 6}); err != nil {
		t.Fatal(err)
	}
	response, err = mostPlayed(ctx, trackscontroller.MostPlayedRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Tracks) != 3 || response.Tracks[1].ID != ids[0] || response.Tracks[2].ID != ids[2] || response.Meta != controllers.NewMeta(3, 1, 10) {
		t.Fatalf("tied most played order: %+v", response)
	}
}
