package filestore

import "io"

// FileStore ...
type FileStore interface {
	Exists(key string) (bool, error)
	Put(key string, rs io.ReadSeeker) error
}
