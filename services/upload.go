package services

import (
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/tombell/memoir/database"
)

const (
	defaultAWSRegion = "us-east-1"
	defaultS3Bucket  = "memoir-uploads"
)

// Upload uploads a new object to S3, reading the bytes from the given Reader.
func (s *Services) Upload(r io.Reader, key, contentType string) (string, error) {
	input := &s3manager.UploadInput{
		Bucket:      aws.String(defaultS3Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        r,
	}

	creds := credentials.NewStaticCredentials(s.Config.AWS.Key, s.Config.AWS.Secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(defaultAWSRegion)

	uploader := s3manager.NewUploader(session.New(cfg))

	result, err := uploader.Upload(input)
	if err != nil {
		return "", errors.Wrap(err, "s3 upload failed")
	}

	return result.Location, nil
}

// AssociateUpload associates an uploaded mix to the given tracklist.
// TODO: return a MixUpload?
func (s *Services) AssociateUpload(filename, location, tracklistName string) error {
	tracklist, err := s.DB.FindTracklist(tracklistName)
	if err != nil {
		return errors.Wrap(err, "find tracklist failed")
	}
	if tracklist == nil {
		return fmt.Errorf("tracklist named %q doesn't exist", tracklistName)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "db begin failed")
	}

	id, _ := uuid.NewV4()

	upload := &database.MixUploadRecord{
		ID:          id.String(),
		TracklistID: tracklist.ID,
		Filename:    filename,
		Location:    location,
		Created:     time.Now().UTC(),
		Updated:     time.Now().UTC(),
	}

	if err := s.DB.InsertMixUpload(tx, upload); err != nil {
		return errors.Wrap(err, "insert mix_upload failed")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "tx commit failed")
	}

	return nil
}
