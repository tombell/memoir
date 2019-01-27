package services

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/tombell/memoir/database"
)

const (
	defaultS3Region = "us-east-1"
	defaultS3Bucket = "memoir-uploads"
)

// UploadMix ...
func (s *Services) UploadMix(file, tracklistName string) (string, error) {
	tracklist, err := s.DB.FindTracklist(tracklistName)
	if err != nil {
		return "", errors.Wrap(err, "find tracklist failed")
	}
	if tracklist == nil {
		return "", fmt.Errorf("tracklist named %q doesn't exist", tracklistName)
	}

	tx, err := s.DB.Begin()
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

	key, err := s.generateObjectKey(bytes.NewBuffer(data), ext)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "generate filename failed")
	}

	id, _ := uuid.NewV4()

	upload := &database.MixUploadRecord{
		ID:          id.String(),
		TracklistID: tracklist.ID,
		Filename:    filename,
		Location:    key,
		Created:     time.Now().UTC(),
		Updated:     time.Now().UTC(),
	}

	if err := s.DB.InsertMixUpload(tx, upload); err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "insert mix_upload failed")
	}

	exists, err := s.uploadExists(key)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "check upload exists failed")
	}

	if !exists {
		contentType, err := s.detectContentType(bytes.NewBuffer(data))
		if err != nil {
			tx.Rollback()
			return "", errors.Wrap(err, "detect content type failed")
		}

		if err := s.upload(bytes.NewBuffer(data), filename, key, contentType); err != nil {
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

func (s *Services) detectContentType(r io.Reader) (string, error) {
	var buf [512]byte

	if _, err := r.Read(buf[:]); err != nil {
		return "", errors.Wrap(err, "read failed")
	}

	return http.DetectContentType(buf[:]), nil
}

func (s *Services) uploadExists(key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(defaultS3Bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-1"),
	}

	creds := credentials.NewStaticCredentials(s.Config.AWS.Key, s.Config.AWS.Secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(defaultS3Region)
	svc := s3.New(session.New(cfg))

	if _, err := svc.GetObject(input); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
		}

		return false, errors.Wrap(err, "get object failed")
	}

	return true, nil
}

func (s *Services) upload(r io.Reader, filename, key, contentType string) error {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(defaultS3Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        r,
		Metadata: map[string]*string{
			"memoir-filename": aws.String(filename),
		},
	}

	creds := credentials.NewStaticCredentials(s.Config.AWS.Key, s.Config.AWS.Secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(defaultS3Region)
	uploader := s3manager.NewUploader(session.New(cfg))

	if _, err := uploader.Upload(input); err != nil {
		return errors.Wrap(err, "s3 upload failed")
	}

	return nil
}
