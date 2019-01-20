package services

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	defaultAWSRegion = "us-east-1"
	defaultS3Bucket  = "memoir-uploads"
)

// UploadFile uploads a single file at the given path to S3.
func (s *Services) UploadFile(filepath string, key string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	input := &s3manager.UploadInput{
		Bucket: aws.String(defaultS3Bucket),
		Key:    aws.String(key),
		Body:   f,
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
