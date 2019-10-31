package services

import (
	"log"

	"github.com/tombell/memoir/config"
	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/filestore"
)

// Services contains shared dependencies required by service methods.
type Services struct {
	Config    *config.Config
	DataStore *datastore.Store
	FileStore filestore.Store
	Logger    *log.Logger
}
