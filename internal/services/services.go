package services

import (
	"log"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/filestore"
)

// Services contains shared dependencies required by service methods.
type Services struct {
	Config    *config.Config
	DataStore *datastore.Store
	FileStore *filestore.Store
	Logger    *log.Logger
}
