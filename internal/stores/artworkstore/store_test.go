package artworkstore

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"testing"

	apierrors "github.com/tombell/memoir/internal/errors"
)

func TestUploadPreservesKeyAndContents(t *testing.T) {
	encoders := map[string]func(io.Writer, image.Image) error{
		"png": png.Encode,
		"gif": func(w io.Writer, img image.Image) error { return gif.Encode(w, img, nil) },
		"jpg": func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) },
	}
	for extension, encode := range encoders {
		t.Run(extension, func(t *testing.T) {
			var data bytes.Buffer
			if err := encode(&data, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
				t.Fatal(err)
			}
			for _, exists := range []bool{false, true} {
				t.Run(fmt.Sprintf("exists=%t", exists), func(t *testing.T) {
					files := &fakeFileStore{exists: exists}
					r := bytes.NewReader(data.Bytes())
					// Hashing and upload must start at the beginning, even for an advanced reader.
					if _, err := r.Seek(1, io.SeekStart); err != nil {
						t.Fatal(err)
					}
					upload, gotExists, err := New(files).Upload(context.Background(), r, "cover."+extension)
					if err != nil {
						t.Fatal(err)
					}
					wantKey := fmt.Sprintf("%x.%s", md5.Sum(data.Bytes()), extension)
					if upload.Key != wantKey || files.key != wantKey || gotExists != exists {
						t.Fatalf("upload=%+v, checked key=%q, exists=%t", upload, files.key, gotExists)
					}
					if exists && files.putCalls != 0 {
						t.Fatal("existing artwork was uploaded again")
					}
					if !exists && (files.putCalls != 1 || !bytes.Equal(files.data, data.Bytes())) {
						t.Fatal("uploaded content differs from the original file")
					}
				})
			}
		})
	}
}

func TestUploadRejectsInvalidFilesBeforeStorage(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		status int
	}{
		{"empty", nil, http.StatusBadRequest},
		{"short text", []byte("x"), http.StatusUnsupportedMediaType},
		{"html named png", []byte("<!DOCTYPE html><html>not artwork</html>"), http.StatusUnsupportedMediaType},
		{"svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`), http.StatusUnsupportedMediaType},
		{"pdf", []byte("%PDF-1.7"), http.StatusUnsupportedMediaType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := &fakeFileStore{}
			_, _, err := New(files).Upload(context.Background(), bytes.NewReader(tt.data), "artwork.png")
			var reported *apierrors.Error
			if !errors.As(err, &reported) || reported.Status() != tt.status || len(reported.Message()["artwork"]) == 0 {
				t.Fatalf("expected artwork error %d, got %v", tt.status, err)
			}
			if files.existsCalls != 0 || files.putCalls != 0 {
				t.Fatal("invalid artwork reached storage")
			}
		})
	}
}

func TestUploadSizeBoundary(t *testing.T) {
	for _, size := range []int{MaxUploadSize, MaxUploadSize + 1} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := make([]byte, size)
			copy(data, "GIF89a")
			files := &fakeFileStore{}
			upload, _, err := New(files).Upload(context.Background(), bytes.NewReader(data), "cover.GIF")
			if size == MaxUploadSize {
				if err != nil || upload.Key != fmt.Sprintf("%x.GIF", md5.Sum(data)) || !bytes.Equal(files.data, data) {
					t.Fatalf("exactly 10 MiB should retain its full hash, extension, and contents: %v", err)
				}
				return
			}
			var reported *apierrors.Error
			if !errors.As(err, &reported) || reported.Status() != http.StatusRequestEntityTooLarge {
				t.Fatalf("expected 413, got %v", err)
			}
			if files.existsCalls != 0 || files.putCalls != 0 {
				t.Fatal("oversized artwork reached storage")
			}
		})
	}
}

func TestUploadPropagatesReadSeekAndStorageErrors(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name         string
		reader       *failingReader
		files        *fakeFileStore
		wantExists   int
		wantPutCalls int
	}{
		{"initial seek", &failingReader{failSeek: 1}, &fakeFileStore{}, 0, 0},
		{"sniff read", &failingReader{failReadAtSeek: 1}, &fakeFileStore{}, 0, 0},
		{"rewind for hash", &failingReader{failSeek: 2}, &fakeFileStore{}, 0, 0},
		{"hash read", &failingReader{failReadAtSeek: 2}, &fakeFileStore{}, 0, 0},
		{"rewind for upload", &failingReader{failSeek: 3}, &fakeFileStore{}, 0, 0},
		{"existence check", &failingReader{}, &fakeFileStore{existsErr: failure}, 1, 0},
		{"upload", &failingReader{}, &fakeFileStore{putErr: failure}, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reader.Reader = bytes.NewReader([]byte("GIF89a"))
			tt.reader.failure = failure
			_, _, err := New(tt.files).Upload(context.Background(), tt.reader, "cover.gif")
			if !errors.Is(err, failure) {
				t.Fatalf("expected original error, got %v", err)
			}
			if tt.files.existsCalls != tt.wantExists || tt.files.putCalls != tt.wantPutCalls {
				t.Fatalf("unexpected storage calls: %+v", tt.files)
			}
		})
	}
}

type fakeFileStore struct {
	exists      bool
	existsErr   error
	putErr      error
	existsCalls int
	putCalls    int
	key         string
	data        []byte
}

func (s *fakeFileStore) Exists(_ context.Context, key string) (bool, error) {
	s.existsCalls++
	s.key = key
	return s.exists, s.existsErr
}

func (s *fakeFileStore) Put(_ context.Context, key string, r io.ReadSeeker) error {
	s.putCalls++
	if key != s.key {
		return fmt.Errorf("put key %q differs from checked key %q", key, s.key)
	}
	var err error
	s.data, err = io.ReadAll(r)
	if err != nil {
		return err
	}
	return s.putErr
}

type failingReader struct {
	*bytes.Reader
	seekCalls      int
	failSeek       int
	failReadAtSeek int
	failure        error
}

func (r *failingReader) Seek(offset int64, whence int) (int64, error) {
	r.seekCalls++
	if r.seekCalls == r.failSeek {
		return 0, r.failure
	}
	return r.Reader.Seek(offset, whence)
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.seekCalls == r.failReadAtSeek {
		return 0, r.failure
	}
	return r.Reader.Read(p)
}
