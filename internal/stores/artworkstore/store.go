package artworkstore

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/tombell/memoir/internal/errors"
)

// MaxUploadSize is the maximum artwork file size in bytes (10 MiB).
const MaxUploadSize = 10 << 20

// FileStore defines the file operations used by the artwork store.
type FileStore interface {
	Exists(context.Context, string) (bool, error)
	Put(context.Context, string, io.ReadSeeker) error
}

// Store is a store used for interacting with a file store for artwork files.
type Store struct {
	fileStore FileStore
}

// New returns a new Store.
func New(store FileStore) *Store {
	return &Store{fileStore: store}
}

// Upload will upload the given file data if it does not already exist in the
// file store. It returns a model containing the "key" of the object.
func (s *Store) Upload(ctx context.Context, r io.ReadSeeker, filename string) (*Upload, bool, error) {
	op := errors.Op("artworkstore[upload]")

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, false, errors.E(op, errors.Strf("seeking file failed: %w", err))
	}

	var buf [512]byte
	n, err := io.ReadFull(r, buf[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, false, errors.E(op, errors.Strf("reading file failed: %w", err))
	}
	if n == 0 {
		return nil, false, errors.E(op, errors.M{"artwork": {"must not be empty"}}, http.StatusBadRequest)
	}

	switch http.DetectContentType(buf[:n]) {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp":
	default:
		return nil, false, errors.E(op, errors.M{"artwork": {"must be a JPEG, PNG, GIF, WebP, or BMP image"}}, http.StatusUnsupportedMediaType)
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, false, errors.E(op, errors.Strf("seeking file failed: %w", err))
	}

	ext := filepath.Ext(filename)

	h := md5.New()

	size, err := io.Copy(h, io.LimitReader(r, MaxUploadSize+1))
	if err != nil {
		return nil, false, errors.E(op, errors.Strf("hashing file failed: %w", err))
	}
	if size > MaxUploadSize {
		return nil, false, errors.E(op, errors.M{"artwork": {"must be at most 10 MiB"}}, http.StatusRequestEntityTooLarge)
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, false, errors.E(op, errors.Strf("seeking file failed: %w", err))
	}

	key := fmt.Sprintf("%x%s", h.Sum(nil), ext)

	exists, err := s.fileStore.Exists(ctx, key)
	if err != nil {
		return nil, false, errors.E(op, errors.Strf("checking if file exists failed: %w", err))
	}

	if !exists {
		if err := s.fileStore.Put(ctx, key, r); err != nil {
			return nil, false, errors.E(op, errors.Strf("putting file failed: %w", err))
		}
	}

	return &Upload{Key: key}, exists, nil
}
