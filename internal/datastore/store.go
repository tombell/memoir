package datastore

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*pgxpool.Pool
	*Queries
}

func NewStore(dbpool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    dbpool,
		Queries: New(dbpool),
	}
}
