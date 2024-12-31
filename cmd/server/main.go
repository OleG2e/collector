package main

import (
	"context"
	"log"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/controller"
	"github.com/OleG2e/collector/internal/storage"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	l, zapErr := logging.NewZapLogger(zap.InfoLevel)

	if zapErr != nil {
		log.Panic(zapErr)
	}

	ctx := context.Background()
	ctx = l.WithContextFields(ctx, zap.String("app", "server"))

	defer l.Sync()

	conf := config.NewServerConfig(ctx, l)

	l.SetLevel(conf.GetLogLevel())

	store := storage.NewMemStorage(ctx, l, conf)

	defer func(storage *storage.MemStorage) {
		if flushErr := storage.FlushStorage(); flushErr != nil {
			l.ErrorCtx(ctx, "flush storage error", zap.Error(flushErr))
		}
	}(store)

	storeInterval := conf.GetStoreIntervalDuration()

	if storeInterval > 0 {
		store.InitFlushStorageTicker(storeInterval)
	}

	if conf.Restore {
		store.RestoreStorage()
	}

	metricsController := controller.New(l, ctx, store, conf)

	if err := metricsController.Routes().ServeHTTP(conf); err != nil {
		l.PanicCtx(ctx, "failed to start server", zap.Error(err))
	}
}
