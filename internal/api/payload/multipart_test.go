package payload

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	apierrors "github.com/tombell/memoir/internal/errors"
)

func TestCleanupAfterPartialMultipartDecoding(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("artwork", "cover.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(make([]byte, (1<<20)+1)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/artwork", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	input, err := Read[struct {
		Artwork *File `file:"artwork"`
		Missing *File `file:"missing"`
	}](r)
	var reported *apierrors.Error
	if !errors.As(err, &reported) || reported.Status() != http.StatusBadRequest || input.Artwork == nil {
		t.Fatalf("expected partial input and a client error, got input=%+v, err=%v", input, err)
	}
	files, err := os.ReadDir(tmp)
	if err != nil || len(files) == 0 {
		t.Fatalf("expected a disk-backed multipart file, got %v, err=%v", files, err)
	}
	if err := Cleanup(r, input); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Artwork.File.Read(make([]byte, 1)); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("opened file was not closed after the decoding error: %v", err)
	}
	files, err = os.ReadDir(tmp)
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary files survived cleanup: %v, err=%v", files, err)
	}
}
