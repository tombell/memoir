package services

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"
)

const (
	bucketArtworkUploads = "memoir-artwork"
)

// UploadArtwork uploads the artwork at the given path to the configured storage
// backend.
func (s *Services) UploadArtwork(r io.ReadSeeker, filename string) (string, error) {
	ext := filepath.Ext(filename)

	key, err := s.generateObjectKey(r, ext)
	if err != nil {
		return "", fmt.Errorf("generate filename failed: %w", err)
	}

	exists, err := s.FileStore.Exists(bucketArtworkUploads, key)
	if err != nil {
		return "", fmt.Errorf("check upload exists failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.FileStore.Put(bucketArtworkUploads, key, r); err != nil {
			return "", fmt.Errorf("filestore put failed: %w", err)
		}
	}

	return key, nil
}

func (s *Services) generateObjectKey(r io.Reader, ext string) (string, error) {
	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("io copy failed: %w", err)
	}

	return fmt.Sprintf("%x%s", h.Sum(nil), ext), nil
}
