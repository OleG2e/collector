package main

import (
	"context"
	"log/slog"
	"net/http"

	"collector/internal/adapters/api/rest"
	"collector/internal/adapters/store"
	"collector/internal/config"
	"collector/internal/controller"
	"collector/internal/core/domain"
	"collector/internal/core/services"
	"collector/pkg/logging"
	"collector/pkg/network"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: log}
		}),
		fx.Supply(domain.NewMetrics()),
		fx.Provide(context.Background),
		fx.Provide(config.NewServerConfig),
		fx.Provide(services.NewStoreService),
		fx.Provide(getStorage),
		fx.Provide(newLogger),
		fx.Provide(network.NewResponse),
		fx.Provide(rest.NewAPI),
		fx.Provide(rest.NewRouter),
		fx.Provide(controller.New),
		fx.Provide(newHTTPServer),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func newHTTPServer(
	lc fx.Lifecycle,
	mux *chi.Mux,
	logger *slog.Logger,
	conf *config.ServerConfig,
	storeService *services.StoreService,
) *http.Server {
	srv := &http.Server{
		Addr:    conf.GetAddress(),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("http server listening on " + srv.Addr)

			go func() {
				if srvErr := srv.ListenAndServe(); srvErr != nil {
					logger.ErrorContext(ctx, "http server start error", slog.Any("error", srvErr))
					if errShutdown := srv.Shutdown(context.Background()); errShutdown != nil {
						logger.ErrorContext(
							ctx,
							"http server shutdown error",
							slog.Any("error", errShutdown),
						)
					}
				}
			}()

			storeInterval := conf.GetStoreIntervalDuration()
			if storeInterval > 0 {
				storeService.InitFlushStorageTicker(ctx, storeInterval)
			}

			if conf.Restore {
				if restoreErr := storeService.Restore(ctx); restoreErr != nil {
					logger.ErrorContext(ctx, "restore storage error", slog.Any("error", restoreErr))

					return restoreErr
				}
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if flushErr := storeService.Save(ctx); flushErr != nil {
				logger.ErrorContext(ctx, "flush storage error", slog.Any("error", flushErr))

				return flushErr
			}
			if closeErr := storeService.Close(); closeErr != nil {
				logger.ErrorContext(ctx, "failed to close storage", slog.Any("error", closeErr))

				return closeErr
			}

			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func newLogger(conf *config.ServerConfig) *slog.Logger {
	return logging.NewLogger(conf.GetLogLevel())
}

func getStorage(
	ctx context.Context,
	logger *slog.Logger,
	conf *config.ServerConfig,
	metrics *domain.Metrics,
) store.Store {
	st := store.NewStore(ctx, logger, conf, metrics)

	logger.DebugContext(ctx, "Using store algo", slog.String("algo", string(st.GetStoreType())))

	return st
}
