package filestore

import (
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const (
	defaultS3Region = "us-east-1"
)

// S3 is a file store implementing the FileStore interface, using S3 as the
// storage backend.
type S3 struct {
	svc    *s3.S3
	bucket string
}

// NewS3 returns an initialised S3 storage layer.
func NewS3(bucket, key, secret string) *S3 {
	creds := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(defaultS3Region)

	return &S3{
		svc:    s3.New(session.New(cfg)),
		bucket: bucket,
	}
}

// Exists checks if the object with the given key exists in the S3 bucket.
func (s *S3) Exists(key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-1"),
	}

	if _, err := s.svc.GetObject(input); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
		}

		return false, errors.Wrap(err, "get object failed")
	}

	return true, nil
}

// Put uploads an object with the given key to the S3 bucket.
func (s *S3) Put(key string, r io.ReadSeeker) error {
	var buf [512]byte

	if _, err := r.Read(buf[:]); err != nil {
		return errors.Wrap(err, "read failed")
	}

	contentType := http.DetectContentType(buf[:])

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        r,
	}

	if _, err := s.svc.PutObject(input); err != nil {
		return errors.Wrap(err, "s3 put object failed")
	}

	return nil
}
