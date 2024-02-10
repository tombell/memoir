package services

import (
	"errors"
)

var (
	ErrQueryFailed     = errors.New("query failed")
	ErrTracklistExists = errors.New("tracklist with name already exists")
)
