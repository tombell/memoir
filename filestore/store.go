package filestore

import "io"

// Store is an interface for uploading objects (files) to a storage backend.
type Store interface {
	Exists(bucket, key string) (bool, error)
	Put(bucket, key string, rs io.ReadSeeker) error
}
