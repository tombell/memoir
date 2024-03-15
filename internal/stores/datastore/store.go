package datastore

import (
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/tombell/memoir/internal/database"
)

type Store struct {
	*pgxpool.Pool
	*db.Queries
}

func New(dbpool *pgxpool.Pool) *Store {
	return &Store{
		Pool:    dbpool,
		Queries: db.New(dbpool),
	}
}
