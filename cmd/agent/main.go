package main

import (
	"context"
	"log"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/internal/storage"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

func main() {
	l, err := logging.NewZapLogger(zap.DebugLevel)

	if err != nil {
		log.Panic(err)
	}

	ctx := context.Background()
	ctx = l.WithContextFields(ctx, zap.String("app", "agent"))

	defer l.Sync()

	agentConfig := config.NewAgentConfig(ctx, l)

	l.SetLevel(agentConfig.GetLogLevel())

	storage.RunMonitor(ctx, l, agentConfig)
}
