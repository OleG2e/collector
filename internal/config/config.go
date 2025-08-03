package config

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	defaultReportIntervalSeconds = 10
	defaultPollIntervalSeconds   = 2
	defaultStoreIntervalSeconds  = 300
)

type (
	AgentConfig struct {
		AppName        string
		LogLevel       string `env:"LOG_LEVEL"`
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
		HashKey        string `env:"KEY"`
		RateLimit      int    `env:"RATE_LIMIT"`
	}
	ServerConfig struct {
		AppName         string
		LogLevel        string `env:"LOG_LEVEL"`
		Address         string `env:"ADDRESS"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DSN             string `env:"DATABASE_DSN"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		Restore         bool   `env:"RESTORE"`
		HashKey         string `env:"KEY"`
	}
)

func NewAgentConfig(ctx context.Context) (*AgentConfig, error) {
	config := AgentConfig{AppName: "agent"}

	err := env.Parse(&config)
	logger := slog.Default()
	logger.InfoContext(ctx, "init agent env config",
		slog.String("LOG_LEVEL", os.Getenv("LOG_LEVEL")),
		slog.String("ADDRESS", os.Getenv("ADDRESS")),
		slog.String("REPORT_INTERVAL", os.Getenv("REPORT_INTERVAL")),
		slog.String("POLL_INTERVAL", os.Getenv("POLL_INTERVAL")),
		slog.String("KEY", os.Getenv("KEY")),
		slog.String("RATE_LIMIT", os.Getenv("RATE_LIMIT")),
	)

	if err != nil {
		return nil, fmt.Errorf("parse env agent config error: %w", err)
	}

	var (
		addr, logLevel, hashKey                 string
		reportInterval, pollInterval, rateLimit int
	)

	flag.StringVar(&logLevel, "log_level", "debug", "log level")
	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.StringVar(&hashKey, "k", "", "hash key")
	flag.IntVar(&reportInterval, "r", defaultReportIntervalSeconds, "report interval")
	flag.IntVar(&pollInterval, "p", defaultPollIntervalSeconds, "poll interval")
	flag.IntVar(&rateLimit, "l", 0, "rate limit")

	flag.Parse()

	logger.InfoContext(ctx, "init agent flags config",
		slog.String("LOG_LEVEL", logLevel),
		slog.String("ADDRESS", addr),
		slog.String("KEY", hashKey),
		slog.Int("REPORT_INTERVAL", reportInterval),
		slog.Int("POLL_INTERVAL", pollInterval),
		slog.Int("RATE_LIMIT", rateLimit),
	)

	if config.LogLevel == "" {
		config.LogLevel = logLevel
	}

	if config.Address == "" {
		config.Address = addr
	}

	if config.HashKey == "" {
		config.HashKey = hashKey
	}

	if config.ReportInterval == 0 {
		config.ReportInterval = reportInterval
	}

	if config.PollInterval == 0 {
		config.PollInterval = pollInterval
	}

	if config.RateLimit == 0 {
		config.RateLimit = rateLimit
	}

	logger.InfoContext(ctx, "final agent params",
		slog.String("ADDRESS", config.Address),
		slog.Int("REPORT_INTERVAL", config.ReportInterval),
		slog.Int("POLL_INTERVAL", config.PollInterval),
		slog.Int("RATE_LIMIT", config.RateLimit),
		slog.String("LOG_LEVEL", config.LogLevel),
		slog.String("KEY", config.HashKey),
	)

	return &config, nil
}

func NewServerConfig(ctx context.Context) (*ServerConfig, error) {
	config := ServerConfig{AppName: "server"}

	err := env.Parse(&config)

	logger := slog.Default()
	logger.InfoContext(ctx, "init server env config",
		slog.String("ADDRESS", os.Getenv("ADDRESS")),
		slog.String("LOG_LEVEL", os.Getenv("LOG_LEVEL")),
		slog.String("STORE_INTERVAL", os.Getenv("STORE_INTERVAL")),
		slog.String("FILE_STORAGE_PATH", os.Getenv("FILE_STORAGE_PATH")),
		slog.String("RESTORE", os.Getenv("RESTORE")),
		slog.String("DATABASE_DSN", os.Getenv("DATABASE_DSN")),
		slog.String("KEY", os.Getenv("KEY")),
	)

	if err != nil {
		return nil, fmt.Errorf("parse env server config error: %w", err)
	}

	var (
		addr, logLevel, fileStorage, dsn, hashKey string
		storeInterval                             int
		restore                                   bool
	)

	flag.StringVar(&addr, "a", "localhost:8080", "server host:port")
	flag.StringVar(&logLevel, "log_level", "debug", "log level")
	flag.StringVar(&hashKey, "k", "", "hash key")
	flag.IntVar(&storeInterval, "i", defaultStoreIntervalSeconds, "store interval")
	flag.StringVar(&fileStorage, "f", "storage.db", "file storage path")
	flag.BoolVar(&restore, "r", true, "restore previous data")
	flag.StringVar(&dsn, "d", "", "postgres DSN")

	flag.Parse()

	logger.InfoContext(ctx, "init server flags config",
		slog.String("ADDRESS", addr),
		slog.String("FILE_STORAGE_PATH", fileStorage),
		slog.String("LOG_LEVEL", logLevel),
		slog.String("DATABASE_DSN", dsn),
		slog.Int("STORE_INTERVAL", storeInterval),
		slog.Bool("RESTORE", restore),
		slog.String("KEY", hashKey),
	)

	if config.LogLevel == "" {
		config.LogLevel = logLevel
	}

	if config.Address == "" {
		config.Address = addr
	}

	if config.HashKey == "" {
		config.HashKey = hashKey
	}

	if config.FileStoragePath == "" {
		config.FileStoragePath = fileStorage
	}

	if config.DSN == "" {
		config.DSN = dsn
	}

	v, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		vInt, vErr := strconv.Atoi(v)
		if vErr != nil {
			return nil, fmt.Errorf("convert STORE_INTERVAL env to int error: %w", vErr)
		}

		config.StoreInterval = vInt
	} else {
		config.StoreInterval = storeInterval
	}

	v, ok = os.LookupEnv("RESTORE")
	if ok {
		vBool, vBoolErr := strconv.ParseBool(v)
		if vBoolErr != nil {
			return nil, fmt.Errorf("convert RESTORE env to bool error: %w", vBoolErr)
		}

		config.Restore = vBool
	} else {
		config.Restore = restore
	}

	logger.InfoContext(ctx, "final server params",
		slog.String("ADDRESS", config.Address),
		slog.String("FILE_STORAGE_PATH", config.FileStoragePath),
		slog.Int("STORE_INTERVAL", config.StoreInterval),
		slog.Bool("RESTORE", config.Restore),
		slog.String("LOG_LEVEL", config.LogLevel),
		slog.String("DATABASE_DSN", config.DSN),
		slog.String("KEY", config.HashKey),
	)

	return &config, nil
}

func (c *AgentConfig) GetLogLevel() slog.Level {
	return parseLogLevel(c.LogLevel)
}

func (c *ServerConfig) GetLogLevel() slog.Level {
	return parseLogLevel(c.LogLevel)
}

func (c *AgentConfig) GetAddress() string {
	return c.Address
}

func (c *AgentConfig) GetHashKey() string {
	return c.HashKey
}

func (c *AgentConfig) HasHashKey() bool {
	return c.HashKey != ""
}

func (c *ServerConfig) GetAddress() string {
	return c.Address
}

func (c *AgentConfig) GetReportIntervalDuration() time.Duration {
	return time.Duration(c.ReportInterval) * time.Second
}

func (c *AgentConfig) GetPollIntervalDuration() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}

func (c *ServerConfig) GetStoreIntervalDuration() time.Duration {
	return time.Duration(c.StoreInterval) * time.Second
}

func (c *ServerConfig) GetStoreInterval() int {
	return c.StoreInterval
}

func (c *ServerConfig) GetDSN() string {
	return c.DSN
}

func (c *ServerConfig) GetHashKey() string {
	return c.HashKey
}

func (c *ServerConfig) HasHashKey() bool {
	return c.HashKey != ""
}

func parseLogLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
