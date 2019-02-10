package services

import (
	"log"

	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/storage"
)

// Config contains any configuration data for the service functions.
type Config struct{}

// Services contains shared functionality required by service functions.
type Services struct {
	Config    *Config
	Logger    *log.Logger
	DataStore *datastore.DataStore
	Storage   storage.ObjectStorage
}
