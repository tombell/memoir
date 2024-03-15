package filestore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	cfg "github.com/tombell/memoir/internal/config"
)

type Store struct {
	config *cfg.Config
	svc    *s3.Client
}

func New(cfg *cfg.Config) *Store {
	creds := credentials.NewStaticCredentialsProvider(cfg.AWS.Key, cfg.AWS.Secret, "")
	awscfg, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(cfg.AWS.Region),
	)

	return &Store{
		config: cfg,
		svc:    s3.NewFromConfig(awscfg),
	}
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.AWS.Bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-1"),
	}

	if _, err := s.svc.GetObject(ctx, input); err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return false, nil
		}

		return false, fmt.Errorf("get object failed: %w", err)
	}

	return true, nil
}

func (s *Store) Put(ctx context.Context, key string, r io.ReadSeeker) error {
	var buf [512]byte

	if _, err := r.Read(buf[:]); err != nil {
		return fmt.Errorf("read failed: %w", err)
	}

	r.Seek(0, io.SeekStart)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.AWS.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(http.DetectContentType(buf[:])),
		Body:        r,
	}

	if _, err := s.svc.PutObject(ctx, input); err != nil {
		return fmt.Errorf("s3 put object failed: %w", err)
	}

	return nil
}
