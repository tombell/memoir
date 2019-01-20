package services

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	defaultAWSRegion = "us-east-1"
	defaultS3Bucket  = "memoir-uploads"
)

// Upload uploads a new object to S3, reading the bytes from the given Reader.
func (s *Services) Upload(r io.Reader, key string, contentType string) (string, error) {
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
		return "", err
	}

	return result.Location, nil
}
