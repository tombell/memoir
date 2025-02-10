package datastore

import (
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/tombell/memoir/internal/database"
)

// Store is a store used for interacting with the database. Query functions are
// generated from SQL Queries by SQLC in the database package.
type Store struct {
	*pgxpool.Pool
	*db.Queries
}

// New returns a new Store that is configured to use the given pool of database
// connections.
func New(dbpool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    dbpool,
		Queries: db.New(dbpool),
	}
}
