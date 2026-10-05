package filestore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	cfg "github.com/tombell/memoir/internal/config"
)

func TestExistsUsesHeadWithPinnedSDK(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantExists bool
		wantErr    bool
	}{
		{"exists", http.StatusOK, "", true, false},
		{"missing object without an error body", http.StatusNotFound, "", false, false},
		{"missing object with NotFound", http.StatusNotFound, "<Error><Code>NotFound</Code></Error>", false, false},
		{"forbidden", http.StatusForbidden, "", false, true},
		{"bad request", http.StatusBadRequest, "", false, true},
		{"server failure", http.StatusInternalServerError, "", false, true},
		{"missing bucket", http.StatusNotFound, "<Error><Code>NoSuchBucket</Code></Error>", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			client := s3.NewFromConfig(aws.Config{
				Region:      "eu-west-1",
				Credentials: credentials.NewStaticCredentialsProvider("test-key", "test-secret", ""),
				Retryer:     func() aws.Retryer { return aws.NopRetryer{} },
				HTTPClient: fakeHTTPClient(func(r *http.Request) (*http.Response, error) {
					calls++
					if r.Method != http.MethodHead || r.URL.Path != "/cover.png" || r.Header.Get("Range") != "" {
						t.Fatalf("expected an un-ranged HEAD for cover.png, got %s %s, range=%q", r.Method, r.URL, r.Header.Get("Range"))
					}
					return &http.Response{
						StatusCode: tt.status,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader(tt.body)),
						Request:    r,
					}, nil
				}),
			})
			store := &Store{config: testConfig(), svc: client}
			exists, err := store.Exists(context.Background(), "cover.png")
			if exists != tt.wantExists || (err != nil) != tt.wantErr {
				t.Fatalf("exists=%t, err=%v", exists, err)
			}
			if calls != 1 {
				t.Fatalf("expected one HEAD, got %d requests", calls)
			}
		})
	}
}

func TestExistsPropagatesTransportErrors(t *testing.T) {
	failure := errors.New("offline")
	client := s3.NewFromConfig(aws.Config{
		Region:      "eu-west-1",
		Credentials: credentials.NewStaticCredentialsProvider("test-key", "test-secret", ""),
		Retryer:     func() aws.Retryer { return aws.NopRetryer{} },
		HTTPClient: fakeHTTPClient(func(*http.Request) (*http.Response, error) {
			return nil, failure
		}),
	})
	store := &Store{config: testConfig(), svc: client}
	exists, err := store.Exists(context.Background(), "cover.png")
	if exists || !errors.Is(err, failure) {
		t.Fatalf("expected transport error, got exists=%t, err=%v", exists, err)
	}
}

func TestPutShortFilesAndPartialReads(t *testing.T) {
	for _, data := range [][]byte{[]byte("GIF89a"), []byte("short text")} {
		t.Run(string(data), func(t *testing.T) {
			client := &fakeS3Client{}
			store := &Store{config: testConfig(), svc: client}
			r := &chunkReader{Reader: bytes.NewReader(data)}
			if err := store.Put(context.Background(), "cover.gif", r); err != nil {
				t.Fatal(err)
			}
			if client.putCalls != 1 || client.contentType != http.DetectContentType(data) || !bytes.Equal(client.data, data) {
				t.Fatalf("put=%+v", client)
			}
		})
	}
}

func TestPutRejectsEmptyFilesAndPropagatesFailures(t *testing.T) {
	failure := errors.New("injected failure")
	tests := []struct {
		name       string
		reader     *errorReader
		putErr     error
		wantCalls  int
		wantSource bool
	}{
		{"empty", &errorReader{Reader: bytes.NewReader(nil)}, nil, 0, false},
		{"initial seek", &errorReader{failSeek: 1}, nil, 0, true},
		{"read", &errorReader{failRead: true}, nil, 0, true},
		{"rewind", &errorReader{failSeek: 2}, nil, 0, true},
		{"S3 failure", &errorReader{}, failure, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.reader.Reader == nil {
				tt.reader.Reader = bytes.NewReader([]byte("GIF89a"))
			}
			tt.reader.failure = failure
			client := &fakeS3Client{putErr: tt.putErr}
			store := &Store{config: testConfig(), svc: client}
			err := store.Put(context.Background(), "cover.gif", tt.reader)
			if err == nil || (tt.wantSource && !errors.Is(err, failure)) || client.putCalls != tt.wantCalls {
				t.Fatalf("err=%v, put calls=%d", err, client.putCalls)
			}
		})
	}
}

func TestNewPropagatesAWSInitializationError(t *testing.T) {
	// Do not read the user's AWS files or use metadata services in constructor tests.
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty")
	if err := os.WriteFile(empty, []byte("[default]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AWS_CONFIG_FILE", empty)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", empty)
	t.Setenv("AWS_PROFILE", "default")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_DEFAULTS_MODE", "standard")
	t.Setenv("AWS_MAX_ATTEMPTS", "invalid")
	store, err := New(testConfig())
	if err == nil || store != nil || !strings.Contains(err.Error(), "loading AWS config failed") {
		t.Fatalf("expected configuration error and no store, got store=%v, err=%v", store, err)
	}

	t.Setenv("AWS_MAX_ATTEMPTS", "1")
	store, err = New(testConfig())
	if err != nil || store == nil {
		t.Fatalf("expected valid offline initialization, got %v", err)
	}
}

func testConfig() *cfg.Config {
	config := &cfg.Config{}
	config.AWS.Bucket = "memoir-test"
	config.AWS.Region = "eu-west-1"
	config.AWS.Key = "test-key"
	config.AWS.Secret = "test-secret"
	return config
}

type fakeHTTPClient func(*http.Request) (*http.Response, error)

func (fn fakeHTTPClient) Do(r *http.Request) (*http.Response, error) {
	return fn(r)
}

type fakeS3Client struct {
	putCalls    int
	putErr      error
	contentType string
	data        []byte
}

func (c *fakeS3Client) HeadObject(context.Context, *s3.HeadObjectInput, ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	panic("unexpected HeadObject")
}

func (c *fakeS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	c.putCalls++
	c.contentType = aws.ToString(input.ContentType)
	var err error
	c.data, err = io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	return &s3.PutObjectOutput{}, c.putErr
}

type chunkReader struct {
	*bytes.Reader
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(p) > 3 {
		p = p[:3]
	}
	return r.Reader.Read(p)
}

type errorReader struct {
	*bytes.Reader
	seekCalls int
	failSeek  int
	failRead  bool
	failure   error
}

func (r *errorReader) Read(p []byte) (int, error) {
	if r.failRead {
		return 0, r.failure
	}
	return r.Reader.Read(p)
}

func (r *errorReader) Seek(offset int64, whence int) (int64, error) {
	r.seekCalls++
	if r.seekCalls == r.failSeek {
		return 0, r.failure
	}
	return r.Reader.Seek(offset, whence)
}
