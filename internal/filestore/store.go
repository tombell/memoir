package filestore

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Store struct {
	svc *s3.S3
}

func New(key, secret, region string) *Store {
	creds := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(region)
	sess, _ := session.NewSession(cfg)

	return &Store{s3.New(sess)}
}

func (s *Store) Exists(bucket, key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-1"),
	}

	if _, err := s.svc.GetObject(input); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return false, nil
			}
		}

		return false, fmt.Errorf("get object failed: %w", err)
	}

	return true, nil
}

func (s *Store) Put(bucket, key string, r io.ReadSeeker) error {
	var buf [512]byte

	if _, err := r.Read(buf[:]); err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(http.DetectContentType(buf[:])),
		Body:        r,
	}

	if _, err := s.svc.PutObject(input); err != nil {
		return fmt.Errorf("s3 put object failed: %w", err)
	}

	return nil
}
