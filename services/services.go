package services

import (
	"log"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/storage"
)

// Services contains shared functionality required by service functions.
type Services struct {
	Config  *Config
	Logger  *log.Logger
	DB      *database.Database
	Storage storage.ObjectStorage
}
