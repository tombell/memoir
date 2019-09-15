package services

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"

	"github.com/tombell/memoir/datastore"
)

const (
	bucketArtworkUploads = "memoir-artwork"
	bucketMixUploads     = "memoir-uploads"
)

// UploadMix uploads the file at the given path to the configured storage
// backend, and associates with an existing tracklist.
func (s *Services) UploadMix(file, tracklistName string) (string, error) {
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

	id, _ := uuid.NewV4()

	upload := &datastore.MixUpload{
		ID:          id.String(),
		TracklistID: tracklist.ID,
		Filename:    filename,
		Location:    key,
		Created:     time.Now().UTC(),
		Updated:     time.Now().UTC(),
	}

	if err := s.DataStore.AddMixUpload(tx, upload); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("insert mix_upload failed: %w", err)
	}

	exists, err := s.FileStore.Exists(bucketMixUploads, key)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("check upload exists failed: %w", err)
	}

	if !exists {
		if err := s.FileStore.Put(bucketMixUploads, key, r); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("uploading failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return "", fmt.Errorf("tx commit failed: %w", err)
	}

	return key, nil
}

// UploadArtwork ...
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

	exists, err := s.FileStore.Exists(bucketMixUploads, key)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("check upload exists failed: %w", err)
	}

	if !exists {
		if err := s.FileStore.Put(bucketArtworkUploads, key, r); err != nil {
			tx.Rollback()
			return "", fmt.Errorf("uploading failed: %w", err)
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
