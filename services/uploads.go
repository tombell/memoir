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
	"github.com/pkg/errors"

	"github.com/tombell/memoir/datastore"
)

const bucketMixUploads = "memoir-uploads"

// UploadMix uploads the file at the given path to the configured storage
// backend, and associates with an existing tracklist.
func (s *Services) UploadMix(file, tracklistName string) (string, error) {
	tracklist, err := s.DataStore.FindTracklistByName(tracklistName)
	if err != nil {
		return "", errors.Wrap(err, "find tracklist failed")
	}
	if tracklist == nil {
		return "", fmt.Errorf("tracklist named %q doesn't exist", tracklistName)
	}

	tx, err := s.DataStore.Begin()
	if err != nil {
		return "", errors.Wrap(err, "db begin failed")
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "read file failed")
	}

	filename := filepath.Base(file)
	ext := filepath.Ext(filename)

	r := bytes.NewReader(data)

	key, err := s.generateObjectKey(r, ext)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "generate filename failed")
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
		return "", errors.Wrap(err, "insert mix_upload failed")
	}

	exists, err := s.FileStore.Exists(bucketMixUploads, key)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "check upload exists failed")
	}

	if !exists {
		if err := s.FileStore.Put(bucketMixUploads, key, r); err != nil {
			tx.Rollback()
			return "", errors.Wrap(err, "uploading failed")
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "tx commit failed")
	}

	return key, nil
}

func (s *Services) generateObjectKey(r io.Reader, ext string) (string, error) {
	h := md5.New()

	if _, err := io.Copy(h, r); err != nil {
		return "", errors.Wrap(err, "io copy failed")
	}

	return fmt.Sprintf("%x%s", h.Sum(nil), ext), nil
}
