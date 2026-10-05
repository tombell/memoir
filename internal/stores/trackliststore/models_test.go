package trackliststore

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/tombell/memoir/internal/errors"
)

func TestAddTracklistValidatesArtwork(t *testing.T) {
	for _, artwork := range []string{"", strings.Repeat("a", 257)} {
		params := validTracklistParams()
		params.Artwork = artwork
		err := params.Validate()
		if len(err["artwork"]) == 0 || len(err["url"]) != 0 {
			t.Fatalf("artwork validation should check artwork independently of the URL: %v", err)
		}
	}

	params := validTracklistParams()
	params.Artwork = strings.Repeat("a", 256)
	params.URL = ""
	err := params.Validate()
	if len(err["url"]) == 0 || len(err["artwork"]) != 0 {
		t.Fatalf("valid artwork should not fail because URL is empty: %v", err)
	}
}

func TestAddTracklistValidatesTrackRowLength(t *testing.T) {
	for _, row := range [][]string{
		nil, {}, {"Name"}, {"Name", "Artist", "128", "Am"},
		{"Name", "Artist", "128", "Am", "House", "extra"},
	} {
		params := validTracklistParams()
		params.Tracks = append(params.Tracks, row)
		err := params.Validate()
		if len(err) != 1 || len(err["tracks[1]"]) == 0 {
			t.Fatalf("row %#v error = %v, want tracks[1] error", row, err)
		}
	}

	params := validTracklistParams()
	params.Tracks = nil
	if err := params.Validate(); len(err["tracks"]) == 0 {
		t.Fatalf("empty tracks should fail validation: %v", err)
	}
}

func TestAddTracklistValidatesTrackFieldBounds(t *testing.T) {
	for _, tc := range []struct {
		field string
		index int
		value string
	}{
		{"name", 0, ""},
		{"name", 0, strings.Repeat("n", 257)},
		{"artist", 1, ""},
		{"artist", 1, strings.Repeat("a", 257)},
		{"key", 3, strings.Repeat("k", 9)},
		{"genre", 4, strings.Repeat("g", 129)},
	} {
		t.Run(tc.field+"/"+tc.value, func(t *testing.T) {
			params := validTracklistParams()
			params.Tracks[0][tc.index] = tc.value
			err := params.Validate()
			field := "tracks[0]." + tc.field
			if len(err) != 1 || len(err[field]) == 0 {
				t.Fatalf("error = %v, want %s error", err, field)
			}
		})
	}
}

func TestAddTracklistRejectsInvalidBPM(t *testing.T) {
	for _, bpm := range []string{"", "fast", "128bpm", "NaN", "+Inf", "-Inf", "Infinity", "1e309", "-1e309"} {
		t.Run(bpm, func(t *testing.T) {
			params := validTracklistParams()
			params.Tracks[0][2] = bpm
			err := params.Validate()
			if len(err) != 1 || len(err["tracks[0].bpm"]) == 0 {
				t.Fatalf("error = %v, want tracks[0].bpm error", err)
			}
		})
	}
}

func TestAddTracklistAcceptsBoundariesAndUnrestrictedKeys(t *testing.T) {
	for _, bpm := range []string{"0", "-1", "128.5", "1.28e2"} {
		t.Run(bpm, func(t *testing.T) {
			params := validTracklistParams()
			params.Name = strings.Repeat("é", 256)
			params.Artwork = strings.Repeat("é", 256)
			params.Tracks = [][]string{
				{strings.Repeat("é", 256), strings.Repeat("é", 256), bpm, "freeform", strings.Repeat("é", 128)},
				{"Track", "Artist", bpm, "", ""},
				{"Track", "Artist", bpm, "C♯ minor", "House"},
			}
			if err := params.Validate(); err != nil {
				t.Fatalf("valid boundary values should be accepted: %v", err)
			}
		})
	}
}

func TestAddTracklistRejectsInvalidRowsBeforeDatabaseAccess(t *testing.T) {
	// No datastore is provided. Invalid rows must return 422 before a transaction
	// starts or any row fields are indexed by AddTracklist.
	store := New(nil)
	for _, tc := range []struct {
		row   []string
		field string
	}{
		{nil, "tracks[0]"},
		{[]string{"Track", "Artist", "invalid", "Am", "House"}, "tracks[0].bpm"},
		{[]string{"Track", "Artist", "NaN", "Am", "House"}, "tracks[0].bpm"},
	} {
		params := validTracklistParams()
		params.Tracks = [][]string{tc.row}
		_, err := store.AddTracklist(context.Background(), params)
		var e *errors.Error
		if !errors.As(err, &e) || e.Status() != http.StatusUnprocessableEntity || len(e.Message()[tc.field]) == 0 {
			t.Fatalf("error = %v, want 422 with field %q", err, tc.field)
		}
	}
}

func validTracklistParams() *AddTracklistParams {
	return &AddTracklistParams{
		Name: "Mix", Date: "2026-10-05T12:00:00Z", URL: "https://example.com/mix",
		Artwork: "cover.jpg", Tracks: [][]string{{"Track", "Artist", "128.5", "Am", "House"}},
	}
}
