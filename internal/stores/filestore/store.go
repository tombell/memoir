package filestore

import (
	"context"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	cfg "github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/errors"
)

// Store is a store used for interacting with AWS S3.
type Store struct {
	config *cfg.Config
	svc    *s3.Client
}

// New returns a new Store configured for the S3 bucket provided in the given
// configuration.
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

// Exists checks if an object with the given key exists in the bucket.
func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	op := errors.Op("filestore[exists]")

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

		return false, errors.E(op, errors.Strf("get object failed: %w", err))
	}

	return true, nil
}

// Put uploads the file as an object with the given key.
func (s *Store) Put(ctx context.Context, key string, r io.ReadSeeker) error {
	op := errors.Op("filestore[put]")

	var buf [512]byte

	if _, err := r.Read(buf[:]); err != nil {
		return errors.E(op, errors.Strf("read file failed: %w", err))
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return errors.E(op, errors.Strf("seek failed: %w", err))
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.AWS.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(http.DetectContentType(buf[:])),
		Body:        r,
	}

	if _, err := s.svc.PutObject(ctx, input); err != nil {
		return errors.E(op, errors.Strf("put object failed: %w", err))
	}

	return nil
}
