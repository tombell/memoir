package services

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

const (
	bucketArtworkUploads = "memoir-artwork"
)

// UploadArtwork uploads the artwork at the given path to the configured storage
// backend, and associates with a tracklist with the given name.
func (s *Services) UploadArtwork(file, tracklistName string) (string, error) {
	tracklist, err := s.DataStore.FindTracklistByName(tracklistName)
	if err != nil {
		return "", fmt.Errorf("find tracklist failed: %w", err)
	}
	if tracklist == nil {
		return "", fmt.Errorf("tracklist named %q doesn't exist", tracklistName)
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return "", fmt.Errorf("db begin failed: %w", err)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("read file failed: %w", err)
	}

	filename := filepath.Base(file)
	ext := filepath.Ext(filename)

	r := bytes.NewReader(data)

	key, err := s.generateObjectKey(r, ext)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("generate filename failed: %w", err)
	}

	exists, err := s.FileStore.Exists(bucketArtworkUploads, key)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("check upload exists failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	if !exists {
		if err := s.FileStore.Put(bucketArtworkUploads, key, r); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("filestore put failed: %w", err)
		}
	}

	if err := s.DataStore.AddArtworkToTracklist(tx, tracklist.ID, key); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("add artwork to tracklist failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("tx commit failed: %w", err)
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
