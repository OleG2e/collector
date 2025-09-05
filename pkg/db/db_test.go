package db

import (
	"context"
	"testing"
	"time"
)

func TestNewPoolConn_InvalidDSN_ReturnsError(t *testing.T) {
	// Используем короткий таймаут, чтобы быстро завершить попытку подключения.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Некорректный/недостижимый DSN.
	dsn := "postgres://user:pass@127.0.0.1:1/invaliddb?sslmode=disable"

	pool, err := NewPoolConn(ctx, dsn)
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatalf("expected error for invalid DSN, got nil")
	}
}

func TestRunMigrations_InvalidDSN_ReturnsError(t *testing.T) {
	// Некорректный/недостижимый DSN заставит миграции вернуть ошибку.
	dsn := "postgres://user:pass@localhost:1/invaliddb?sslmode=disable"

	if err := runMigrations(dsn); err == nil {
		t.Fatalf("expected error from runMigrations for invalid DSN, got nil")
	}
}
