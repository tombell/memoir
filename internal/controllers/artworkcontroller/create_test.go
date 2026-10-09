package artworkcontroller

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/tombell/memoir/internal/api/payload"
	apierrors "github.com/tombell/memoir/internal/errors"
)

func TestCreateRejectsMissingArtworkWithoutPanic(t *testing.T) {
	tests := map[string]*payload.File{
		"missing artwork": nil,
		"missing file":    {Header: &multipart.FileHeader{Filename: "cover.png"}},
		"missing header":  {File: &testFile{bytes.NewReader(nil)}},
	}
	for name, file := range tests {
		t.Run(name, func(t *testing.T) {
			response, err := Create(nil)(context.Background(), CreateRequest{Artwork: file})
			var reported *apierrors.Error
			if response != nil || !errors.As(err, &reported) || reported.Status() != http.StatusBadRequest || len(reported.Message()["artwork"]) == 0 {
				t.Fatalf("expected a structured client error, got response=%v, err=%v", response, err)
			}
		})
	}
}

type testFile struct {
	*bytes.Reader
}

func (*testFile) Close() error { return nil }
