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
	svc    s3Client
}

type s3Client interface {
	HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// New returns a new Store configured for the S3 bucket provided in the given
// configuration.
func New(cfg *cfg.Config) (*Store, error) {
	op := errors.Op("filestore[new]")

	creds := credentials.NewStaticCredentialsProvider(cfg.AWS.Key, cfg.AWS.Secret, "")
	awscfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(cfg.AWS.Region),
	)
	if err != nil {
		return nil, errors.E(op, errors.Strf("loading AWS config failed: %w", err))
	}

	return &Store{
		config: cfg,
		svc:    s3.NewFromConfig(awscfg),
	}, nil
}

// Exists checks if an object with the given key exists in the bucket.
func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	op := errors.Op("filestore[exists]")

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.AWS.Bucket),
		Key:    aws.String(key),
	}

	if _, err := s.svc.HeadObject(ctx, input); err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}

		return false, errors.E(op, errors.Strf("head object failed: %w", err))
	}

	return true, nil
}

// Put uploads the file as an object with the given key.
func (s *Store) Put(ctx context.Context, key string, r io.ReadSeeker) error {
	op := errors.Op("filestore[put]")

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return errors.E(op, errors.Strf("seek failed: %w", err))
	}

	var buf [512]byte

	n, err := io.ReadFull(r, buf[:])
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return errors.E(op, errors.Strf("read file failed: %w", err))
	}
	if n == 0 {
		return errors.E(op, errors.Strf("file is empty"))
	}

	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return errors.E(op, errors.Strf("seek failed: %w", err))
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.AWS.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(http.DetectContentType(buf[:n])),
		Body:        r,
	}

	if _, err := s.svc.PutObject(ctx, input); err != nil {
		return errors.E(op, errors.Strf("put object failed: %w", err))
	}

	return nil
}
