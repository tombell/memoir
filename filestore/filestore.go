package filestore

import "io"

// FileStore is an interface for uploading objects (files) to a storage backend.
type FileStore interface {
	Exists(bucket, key string) (bool, error)
	Put(bucket, key string, rs io.ReadSeeker) error
}
