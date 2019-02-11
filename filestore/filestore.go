package filestore

import "io"

// FileStore is an interface for uploading objects (files) to a storage backend.
type FileStore interface {
	Exists(key string) (bool, error)
	Put(key string, rs io.ReadSeeker) error
}
