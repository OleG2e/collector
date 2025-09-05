package main

import (
	"context"
	"log/slog"

	"collector/internal/config"
	"collector/internal/core/services"
	"collector/pkg/logging"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: log}
		}),
		fx.Provide(parseEnvs),
		fx.Provide(parseFlags),
		fx.Provide(context.Background),
		fx.Provide(config.NewAgentConfig),
		fx.Provide(services.NewMonitor),
		fx.Provide(newLogger),
		fx.Invoke(runMonitor),
	).Run()
}

func runMonitor(ctx context.Context, monitor *services.Monitor) error {
	return monitor.Run(ctx)
}

func newLogger(conf *config.AgentConfig) *slog.Logger {
	return logging.NewLogger(conf.GetLogLevel())
}

func parseEnvs() *config.EnvContainer {
	conf := new(config.EnvContainer)
	conf.Parse()

	return conf
}

func parseFlags() *config.FlagContainer {
	conf := new(config.FlagContainer)
	conf.AppType = config.AppTypeAgent
	conf.Parse()

	return conf
}
