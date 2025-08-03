package services

import (
	"context"
	"log/slog"
	"time"

	"collector/internal/adapters/store"
	"collector/internal/config"
	"collector/pkg/retry"
)

type StoreService struct {
	conf   *config.ServerConfig
	logger *slog.Logger
	store  store.Store
}

func NewStoreService(
	logger *slog.Logger,
	conf *config.ServerConfig,
	store store.Store,
) *StoreService {
	return &StoreService{
		conf:   conf,
		logger: logger,
		store:  store,
	}
}

func (ss *StoreService) InitFlushStorageTicker(ctx context.Context, storeInterval time.Duration) {
	ticker := time.NewTicker(storeInterval)
	go func() {
		for range ticker.C {
			if err := ss.Save(ctx); err != nil {
				ss.logger.ErrorContext(ctx, "save storage error", slog.Any("error", err))
			}
		}
	}()
}

func (ss *StoreService) Restore(ctx context.Context) error {
	tryErr := retry.Try(func() error {
		return ss.store.Restore(ctx)
	})

	return tryErr
}

func (ss *StoreService) Close() error {
	return retry.Try(func() error {
		return ss.store.Close()
	})
}

func (ss *StoreService) Save(ctx context.Context) error {
	return retry.Try(func() error {
		return ss.store.Save(ctx)
	})
}
