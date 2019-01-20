package services

import (
	"bytes"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	defaultAWSRegion = "us-east-1"
	defaultS3Bucket  = "memoir-uploads"
)

// UploadFile uploads a single file at the given path to S3.
func (s *Services) UploadFile(filepath string, key string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	buf := make([]byte, info.Size())

	if _, err := f.Read(buf); err != nil {
		return nil
	}

	object := &s3.PutObjectInput{
		Bucket:        aws.String(defaultS3Bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buf),
		ContentLength: aws.Int64(info.Size()),
		ContentType:   aws.String(http.DetectContentType(buf)),
	}

	creds := credentials.NewStaticCredentials(s.Config.AWS.Key, s.Config.AWS.Secret, "")
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(defaultAWSRegion)
	client := s3.New(session.New(), cfg)

	if _, err := client.PutObject(object); err != nil {
		return err
	}

	return nil
}
