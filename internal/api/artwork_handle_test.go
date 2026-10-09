package api

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/payload"
)

func TestArtworkHandlerClosesFilesOnSuccessErrorAndPanic(t *testing.T) {
	for _, outcome := range []string{"success", "error", "panic"} {
		t.Run(outcome, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			r := artworkRequest(t, "artwork", "cover.png", paddedPNG(t, (1<<20)+1), "image/png")
			var opened multipart.File
			type input struct {
				Artwork *payload.File `file:"artwork"`
			}
			type output struct {
				OK bool `json:"ok"`
			}
			handler := rw(func(_ context.Context, in input) (*output, error) {
				opened = in.Artwork.File
				files, err := os.ReadDir(tmp)
				if err != nil || len(files) == 0 {
					t.Fatalf("expected a temporary file during the action: %v, err=%v", files, err)
				}
				switch outcome {
				case "error":
					return nil, errors.New("action failed")
				case "panic":
					panic("action panicked")
				}
				return &output{OK: true}, nil
			})
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			w := httptest.NewRecorder()
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				ware.Logger(logger)(handler).ServeHTTP(w, r)
			}()
			if (recovered != nil) != (outcome == "panic") {
				t.Fatalf("unexpected recovered panic: %v", recovered)
			}
			if outcome == "success" && w.Code != http.StatusOK || outcome == "error" && w.Code != http.StatusInternalServerError {
				t.Fatalf("unexpected status %d: %s", w.Code, w.Body)
			}
			if opened == nil {
				t.Fatal("action was not called")
			}
			if _, err := opened.Read(make([]byte, 1)); !errors.Is(err, os.ErrClosed) {
				t.Fatalf("file was not closed: %v", err)
			}
			assertNoMultipartFiles(t, tmp)
		})
	}
}

func TestArtworkHandlerCleansUpOnEarlyDecodingError(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	r := artworkRequest(t, "artwork", "cover.png", paddedPNG(t, (1<<20)+1), "image/png")
	type input struct {
		Artwork *payload.File `file:"artwork"`
		Missing *payload.File `file:"missing"`
	}
	type output struct {
		OK bool `json:"ok"`
	}
	called := false
	handler := rw(func(context.Context, input) (*output, error) {
		called = true
		return &output{OK: true}, nil
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()
	ware.Logger(logger)(handler).ServeHTTP(w, r)
	if called || w.Code != http.StatusBadRequest {
		t.Fatalf("expected a decoding error before the action, got called=%t, status=%d", called, w.Code)
	}
	assertNoMultipartFiles(t, tmp)
}
