package storage

import "io"

// ObjectStorage is an interface for uploading objects (files) to a storage
// layer.
type ObjectStorage interface {
	Exists(key string) (bool, error)
	Put(key string, r io.ReadSeeker) error
}
