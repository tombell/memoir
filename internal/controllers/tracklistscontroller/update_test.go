package tracklistscontroller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/trackliststore"
)

func TestUpdateRejectsInvalidMetadataBeforeDatabaseAccess(t *testing.T) {
	for _, body := range []string{`{"name":null}`, `{"artwork":""}`, `{"date":"invalid"}`, `{"url":false}`, `{"name":"Valid","tracks":[["short"]]}`, `{"tracks":[]}`, `{"tracks":null}`} {
		t.Run(body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/tracklists/mix", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", "00000000-0000-0000-0000-000000000001")
			input, err := payload.Read[UpdateRequest](req)
			if err != nil {
				t.Fatal(err)
			}
			// A nil database makes any accidental access fail the test.
			_, err = Update(trackliststore.New(nil))(context.Background(), input)
			var e *errors.Error
			if !errors.As(err, &e) || e.Status() != http.StatusUnprocessableEntity {
				t.Fatalf("want 422, got %v", err)
			}
		})
	}
}
