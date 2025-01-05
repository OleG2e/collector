package db

import (
	"context"

	"github.com/OleG2e/collector/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPoolConn(ctx context.Context, c *config.ServerConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), c.GetDSN())

	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	return pool, err
}
