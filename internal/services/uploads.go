package services

import (
	"crypto/md5"
	"fmt"
	"io"
	"path/filepath"

	"github.com/tombell/memoir/internal/services/models"
)

func (s *Services) UploadArtwork(r io.ReadSeeker, filename string) (*models.UploadedItem, error) {
	s.Logger.Info("upload-artwork", "filename", filename)

	ext := filepath.Ext(filename)

	key, err := s.generateObjectKey(r, ext)
	if err != nil {
		return nil, fmt.Errorf("generate filename failed: %w", err)
	}

	exists, err := s.FileStore.Exists(s.Config.AWS.Bucket, key)
	if err != nil {
		return nil, fmt.Errorf("check upload exists failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.FileStore.Put(s.Config.AWS.Bucket, key, r); err != nil {
			return nil, fmt.Errorf("filestore put failed: %w", err)
		}
	}

	return &models.UploadedItem{Key: key}, nil
}

func (s *Services) generateObjectKey(r io.Reader, ext string) (string, error) {
	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("io copy failed: %w", err)
	}

	return fmt.Sprintf("%x%s", h.Sum(nil), ext), nil
}
