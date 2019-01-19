package services

import (
	"log"

	"github.com/tombell/memoir/database"
)

// Services ...
type Services struct {
	Logger *log.Logger
	DB     *database.Database
}
