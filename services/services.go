package services

import (
	"log"

	"github.com/tombell/memoir/database"
)

// Services contains shared functionality required by service functions.
type Services struct {
	Logger *log.Logger
	DB     *database.Database
}
