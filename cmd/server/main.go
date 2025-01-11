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
	l, zapErr := logging.NewZapLogger(zap.DebugLevel)

	if zapErr != nil {
		log.Panic(zapErr)
	}

	ctx := context.Background()
	ctx = l.WithContextFields(ctx, zap.String("app", "server"))

	defer l.Sync()

	conf := config.NewServerConfig(ctx, l)

	l.SetLevel(conf.GetLogLevel())

	storeAlgo := storage.GetStoreAlgo(ctx, l, conf)

	l.DebugCtx(ctx, "Using store algo", zap.String("algo", string(storeAlgo.GetStoreType())))

	store := storage.NewMemStorage(ctx, l, conf, storeAlgo)

	defer func(storage *storage.MemStorage) {
		if flushErr := storage.FlushStorage(); flushErr != nil {
			l.ErrorCtx(ctx, "flush storage error", zap.Error(flushErr))
		}
		if err := store.CloseStorage(); err != nil {
			l.PanicCtx(ctx, "failed to close storage", zap.Error(err))
		}
	}(store)

	storeInterval := conf.GetStoreIntervalDuration()

	if storeInterval > 0 {
		store.InitFlushStorageTicker(storeInterval)
	}

	if conf.Restore {
		restoreErr := store.RestoreStorage()
		if restoreErr != nil {
			l.ErrorCtx(ctx, "restore storage error", zap.Error(restoreErr))
		}
	}

	metricsController := controller.New(l, ctx, store, conf)

	if err := metricsController.Routes().ServeHTTP(conf); err != nil {
		l.PanicCtx(ctx, "failed to start server", zap.Error(err))
	}
}
