package datastore

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/database"
)

type Store struct {
	*pgxpool.Pool
	*database.Queries
}

func New(dbpool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    dbpool,
		Queries: database.New(dbpool),
	}
}
