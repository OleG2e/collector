package main

import (
	"context"
	"log"

	"github.com/OleG2e/collector/pkg/db"

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

	poolConn, dbErr := db.NewPoolConn(ctx, conf)
	if poolConn != nil {
		defer poolConn.Close()
	}

	if dbErr != nil {
		l.WarnCtx(ctx, "Unable to connect to database", zap.Error(dbErr))
	}

	storeAlgo := storage.GetStoreAlgo(ctx, l, conf, poolConn)

	l.DebugCtx(ctx, "Using store algo", zap.Int("algo", int(storeAlgo.GetStoreType())))

	store := storage.NewMemStorage(ctx, l, conf, storeAlgo)

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
