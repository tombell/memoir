package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
)

func TestArtworkUploadResponses(t *testing.T) {
	data := artworkPNG(t)
	for _, exists := range []bool{false, true} {
		t.Run(fmt.Sprintf("exists=%t", exists), func(t *testing.T) {
			files := &artworkFiles{exists: exists}
			handler := artworkAPI(files)
			r := artworkRequest(t, "artwork", "cover.PnG", data, "text/plain")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			wantStatus := http.StatusCreated
			if exists {
				wantStatus = http.StatusOK
			}
			if w.Code != wantStatus || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d, body=%s", w.Code, w.Body)
			}
			var response struct {
				Data artworkstore.Upload `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			wantKey := fmt.Sprintf("%x.PnG", md5.Sum(data))
			if response.Data.Key != wantKey || files.key != wantKey {
				t.Fatalf("expected key %q, got response=%q, storage=%q", wantKey, response.Data.Key, files.key)
			}
			if exists && files.putCalls != 0 {
				t.Fatal("duplicate artwork was uploaded")
			}
			if !exists && (files.putCalls != 1 || !bytes.Equal(files.data, data)) {
				t.Fatal("uploaded content differs from request")
			}
		})
	}
}

func TestArtworkUploadClientErrorsAndTemporaryFileCleanup(t *testing.T) {
	tests := []struct {
		name   string
		make   func(*testing.T) *http.Request
		status int
	}{
		{"no body", func(t *testing.T) *http.Request {
			return httptest.NewRequest(http.MethodPost, "/artwork", nil)
		}, http.StatusBadRequest},
		{"missing artwork after a disk-backed file", func(t *testing.T) *http.Request {
			return artworkRequest(t, "other", "other.png", paddedPNG(t, (1<<20)+1), "image/png")
		}, http.StatusBadRequest},
		{"missing boundary", func(t *testing.T) *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/artwork", bytes.NewBufferString("malformed"))
			r.Header.Set("Content-Type", "multipart/form-data")
			return r
		}, http.StatusBadRequest},
		{"malformed body after a disk-backed file", func(t *testing.T) *http.Request {
			r := artworkRequest(t, "artwork", "cover.png", paddedPNG(t, (1<<20)+1), "image/png")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			// Remove the final multipart boundary, after enough data to spill to disk.
			r.Body = io.NopCloser(bytes.NewReader(body[:len(body)-100]))
			r.ContentLength = int64(len(body) - 100)
			return r
		}, http.StatusBadRequest},
		{"wrong content type", func(t *testing.T) *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/artwork", bytes.NewBufferString("not json"))
			r.Header.Set("Content-Type", "application/json")
			return r
		}, http.StatusBadRequest},
		{"empty artwork", func(t *testing.T) *http.Request {
			return artworkRequest(t, "artwork", "cover.png", nil, "image/png")
		}, http.StatusBadRequest},
		{"non-image with an image header", func(t *testing.T) *http.Request {
			return artworkRequest(t, "artwork", "cover.png", []byte("not an image"), "image/png")
		}, http.StatusUnsupportedMediaType},
		{"file over 10 MiB", func(t *testing.T) *http.Request {
			return artworkRequest(t, "artwork", "cover.png", paddedPNG(t, artworkstore.MaxUploadSize+1), "image/png")
		}, http.StatusRequestEntityTooLarge},
		{"whole request over its limit", func(t *testing.T) *http.Request {
			return artworkRequest(t, "artwork", "cover.png", paddedPNG(t, maxArtworkRequestSize+1), "image/png")
		}, http.StatusRequestEntityTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			files := &artworkFiles{}
			r := tt.make(t)
			r.Header.Set("API-Token", "test-token")
			w := httptest.NewRecorder()
			artworkAPI(files).ServeHTTP(w, r)
			if w.Code != tt.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("expected %d JSON, got %d: %s", tt.status, w.Code, w.Body)
			}
			var response struct {
				Errors map[string][]string `json:"errors"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || len(response.Errors["artwork"]) == 0 {
				t.Fatalf("expected a structured artwork error, got %s, err=%v", w.Body, err)
			}
			if files.existsCalls != 0 || files.putCalls != 0 {
				t.Fatal("invalid request reached storage")
			}
			assertNoMultipartFiles(t, tmp)
		})
	}
}

func TestArtworkUploadAccepts10MiBAndCleansUpOnStorageFailure(t *testing.T) {
	for _, failure := range []error{nil, errors.New("storage unavailable")} {
		t.Run(fmt.Sprintf("storage failure=%t", failure != nil), func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			files := &artworkFiles{existsErr: failure}
			data := paddedPNG(t, artworkstore.MaxUploadSize)
			r := artworkRequest(t, "artwork", "cover.png", data, "image/png")
			w := httptest.NewRecorder()
			artworkAPI(files).ServeHTTP(w, r)
			wantStatus := http.StatusCreated
			if failure != nil {
				wantStatus = http.StatusInternalServerError
				if files.putCalls != 0 {
					t.Fatal("uploaded artwork after a failed existence check")
				}
			}
			if w.Code != wantStatus {
				t.Fatalf("expected %d, got %d: %s", wantStatus, w.Code, w.Body)
			}
			assertNoMultipartFiles(t, tmp)
		})
	}
}

func artworkAPI(files *artworkFiles) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{}
	cfg.API.Token = "test-token"
	return New(logger, cfg, nil, nil, artworkstore.New(files)).router
}

func artworkRequest(t *testing.T, field, filename string, data []byte, contentType string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, field, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/artwork", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("API-Token", "test-token")
	return r
}

func artworkPNG(t *testing.T) []byte {
	t.Helper()
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func paddedPNG(t *testing.T, size int) []byte {
	t.Helper()
	data := make([]byte, size)
	copy(data, artworkPNG(t))
	return data
}

func assertNoMultipartFiles(t *testing.T, dir string) {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("multipart temporary files leaked: %v", files)
	}
}

type artworkFiles struct {
	exists      bool
	existsErr   error
	existsCalls int
	putCalls    int
	key         string
	data        []byte
}

func (s *artworkFiles) Exists(_ context.Context, key string) (bool, error) {
	s.existsCalls++
	s.key = key
	return s.exists, s.existsErr
}

func (s *artworkFiles) Put(_ context.Context, _ string, r io.ReadSeeker) error {
	s.putCalls++
	var err error
	s.data, err = io.ReadAll(r)
	return err
}
