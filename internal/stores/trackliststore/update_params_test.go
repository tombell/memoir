package trackliststore

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func decodeUpdate(t *testing.T, body string) UpdateTracklistParams {
	t.Helper()
	var params UpdateTracklistParams
	if err := json.Unmarshal([]byte(body), &params); err != nil {
		t.Fatal(err)
	}
	return params
}

func TestUpdateMetadataPresence(t *testing.T) {
	params := decodeUpdate(t, `{"name":"New mix","artwork":"new-cover.jpg"}`)
	if !params.Name.Present || params.Name.Value != "New mix" || !params.Artwork.Present || params.Date.Present || params.URL.Present {
		t.Fatalf("decoded fields: %+v", params)
	}
	if err := params.Validate(); err != nil {
		t.Fatal(err)
	}
	database := params.ToDatabaseParams("mix-id")
	if !database.UpdateName || !database.UpdateArtwork || database.UpdateDate || database.UpdateUrl || database.Name != "New mix" || database.Artwork != "new-cover.jpg" || database.ID != "mix-id" {
		t.Fatalf("database fields: %+v", database)
	}
	params = decodeUpdate(t, `{"date":"2026-10-05T12:34:56Z","url":"https://example.com/mix"}`)
	if err := params.Validate(); err != nil {
		t.Fatal(err)
	}
	database = params.ToDatabaseParams("mix-id")
	if !database.UpdateDate || !database.UpdateUrl || !database.Date.Equal(time.Date(2026, 10, 5, 12, 34, 56, 0, time.UTC)) {
		t.Fatalf("database fields: %+v", database)
	}
	if err := (&UpdateTracklistParams{}).Validate(); err != nil {
		t.Fatalf("empty patch should retain metadata: %v", err)
	}
}

func TestUpdateMetadataValidation(t *testing.T) {
	for _, field := range []string{"name", "date", "url", "artwork"} {
		for _, value := range []string{`null`, `""`, `"   "`, `42`, `false`, `[]`, `{}`, `"\u0000"`} {
			t.Run(field+"/"+value, func(t *testing.T) {
				params := decodeUpdate(t, fmt.Sprintf(`{%q:%s}`, field, value))
				if err := params.Validate(); err == nil || len(err[field]) == 0 || len(err) != 1 {
					t.Fatalf("want only %s validation error, got %v", field, err)
				}
			})
		}
	}
	for _, tt := range []struct{ field, value string }{
		{"name", strings.Repeat("é", 257)},
		{"artwork", strings.Repeat("a", 257)},
		{"url", "www.example.com"},
		{"date", "2026-10-05"},
		{"date", "2026-02-30T00:00:00Z"},
	} {
		params := decodeUpdate(t, fmt.Sprintf(`{%q:%q}`, tt.field, tt.value))
		if err := params.Validate(); err == nil || len(err[tt.field]) == 0 {
			t.Fatalf("want %s validation error, got %v", tt.field, err)
		}
	}
	params := decodeUpdate(t, fmt.Sprintf(`{"name":%q}`, strings.Repeat("é", 256)))
	if err := params.Validate(); err != nil {
		t.Fatalf("256 characters should be allowed: %v", err)
	}
}
