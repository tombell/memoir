package services

import (
	"log"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/filestore"
)

// Services contains shared dependencies required by service methods.
type Services struct {
	Config    *config.Config
	DataStore *datastore.Store
	FileStore filestore.Store
	Logger    *log.Logger
}
