package store

import (
	"context"
	"log/slog"

	"collector/internal/config"
	"collector/internal/core/domain"
)

type Type string

const (
	FileStoreType   = Type("file")
	DBStoreType     = Type("db")
	MemoryStoreType = Type("memory")
)

type Store interface {
	GetMetrics() *domain.Metrics
	SetMetrics(metrics *domain.Metrics)
	Save(ctx context.Context) error
	Restore(ctx context.Context) error
	Close() error
	GetStoreType() Type
}

func NewStore(
	ctx context.Context,
	logger *slog.Logger,
	conf *config.ServerConfig,
	metrics *domain.Metrics,
) Store {
	dbStorage, dbErr := NewDBStorage(ctx, logger, conf, metrics)
	if dbErr != nil {
		logger.WarnContext(
			ctx,
			"GetStoreAlgo: failed to connect to database",
			slog.Any("error", dbErr),
		)

		fsStorage, fsStorageErr := NewFileStorage(logger, conf, metrics)
		if fsStorageErr != nil {
			logger.WarnContext(
				ctx,
				"GetStoreAlgo: failed to create fs storage",
				slog.Any("error", fsStorageErr),
			)

			return NewMemoryStorage(metrics)
		}

		return fsStorage
	}

	return dbStorage
}
