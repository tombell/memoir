package services

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
)

func (s *Services) UploadArtwork(r io.ReadSeeker, filename string) (*UploadedItem, bool, error) {
	ext := filepath.Ext(filename)

	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return nil, false, fmt.Errorf("io copy failed: %w", err)
	}

	key := fmt.Sprintf("%x%s", h.Sum(nil), ext)

	exists, err := s.FileStore.Exists(s.Config.AWS.Bucket, key)
	if err != nil {
		return nil, false, fmt.Errorf("filestore exists failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.FileStore.Put(s.Config.AWS.Bucket, key, r); err != nil {
			return nil, false, fmt.Errorf("filestore put failed: %w", err)
		}
	}

	return &UploadedItem{Key: key}, exists, nil
}
