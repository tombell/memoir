package artworkstore

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"

	"github.com/tombell/memoir/internal/errors"
	"github.com/tombell/memoir/internal/stores/filestore"
)

// Store is a store used for interacting with a file store for artwork files.
type Store struct {
	fileStore *filestore.Store
}

// New returns a new Store.
func New(store *filestore.Store) *Store {
	return &Store{fileStore: store}
}

// Upload will upload the given file data if it does not already exist in the
// file store. It returns a model containing the "key" of the object.
func (s *Store) Upload(ctx context.Context, r io.ReadSeeker, filename string) (*Upload, bool, error) {
	op := errors.Op("artworkstore[upload]")

	ext := filepath.Ext(filename)

	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return nil, false, errors.E(op, errors.Strf("io copy filed: %w", err))
	}

	key := fmt.Sprintf("%x%s", h.Sum(nil), ext)

	exists, err := s.fileStore.Exists(ctx, key)
	if err != nil {
		return nil, false, errors.E(op, errors.Strf("checking if file exists failed: %w", err))
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.fileStore.Put(ctx, key, r); err != nil {
			return nil, false, errors.E(op, errors.Strf("putting file failed: %w", err))
		}
	}

	return &Upload{Key: key}, exists, nil
}
