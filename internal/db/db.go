package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), dsn)
}
