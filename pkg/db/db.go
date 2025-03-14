package db

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/OleG2e/collector/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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

	if err := runMigrations(c.GetDSN()); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	return pool, err
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
