package db

import (
	"context"
	"os"

	"github.com/OleG2e/collector/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPoolConn(ctx context.Context, c *config.ServerConfig) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, c.GetDSN())

	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	if err != nil {
		return nil, err
	}

	err = initDB(ctx, pool)

	return pool, err
}

func initDB(ctx context.Context, pollConn *pgxpool.Pool) error {
	sqlCommands, err := os.ReadFile("../../init.sql")

	if err != nil {
		return err
	}

	_, err = pollConn.Exec(ctx, string(sqlCommands))

	return err
}
