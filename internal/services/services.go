package services

import (
	"github.com/charmbracelet/log"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/filestore"
)

type Services struct {
	Config    *config.Config
	DataStore *datastore.Store
	FileStore *filestore.Store
	Logger    log.Logger
}
