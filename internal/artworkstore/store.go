package artworkstore

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"

	"github.com/tombell/memoir/internal/filestore"
)

type UploadedItem struct {
	Key string `json:"key"`
}

type Store struct {
	fileStore *filestore.Store
}

func New(store *filestore.Store) *Store {
	return &Store{fileStore: store}
}

func (s *Store) Upload(r io.ReadSeeker, filename string) (*UploadedItem, bool, error) {
	ext := filepath.Ext(filename)

	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return nil, false, fmt.Errorf("io copy failed: %w", err)
	}

	key := fmt.Sprintf("%x%s", h.Sum(nil), ext)

	exists, err := s.fileStore.Exists(key)
	if err != nil {
		return nil, false, fmt.Errorf("filestore exists failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.fileStore.Put(key, r); err != nil {
			return nil, false, fmt.Errorf("filestore put failed: %w", err)
		}
	}

	return &UploadedItem{Key: key}, exists, nil
}
