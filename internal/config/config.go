package config

import (
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
	defaultRateLimit             = 5

	AppTypeServer = AppType("server")
	AppTypeAgent  = AppType("agent")
)

type (
	AppType    string
	BaseConfig struct {
		AppType  AppType
		LogLevel string `env:"LOG_LEVEL"`
		Address  string `env:"ADDRESS"`
		HashKey  string `env:"KEY"`
	}
	AgentConfig struct {
		BaseConfig
		ReportInterval int `env:"REPORT_INTERVAL"`
		PollInterval   int `env:"POLL_INTERVAL"`
		RateLimit      int `env:"RATE_LIMIT"`
	}
	ServerConfig struct {
		BaseConfig
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DSN             string `env:"DATABASE_DSN"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		Restore         bool   `env:"RESTORE"`
	}
	EnvContainer struct {
		AppType         AppType
		LogLevel        string `env:"LOG_LEVEL"`
		Address         string `env:"ADDRESS"`
		HashKey         string `env:"KEY"`
		ReportInterval  int    `env:"REPORT_INTERVAL"`
		PollInterval    int    `env:"POLL_INTERVAL"`
		RateLimit       int    `env:"RATE_LIMIT"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		DSN             string `env:"DATABASE_DSN"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		Restore         bool   `env:"RESTORE"`
	}
	FlagContainer struct {
		AppType         AppType
		LogLevel        string
		Address         string
		HashKey         string
		ReportInterval  int
		PollInterval    int
		RateLimit       int
		FileStoragePath string
		DSN             string
		StoreInterval   int
		Restore         bool
	}
)

func (fc *FlagContainer) Parse() {
	if fc.AppType == AppTypeServer {
		flag.StringVar(&fc.FileStoragePath, "f", "storage.db", "file storage path")
		flag.BoolVar(&fc.Restore, "r", true, "restore previous data")
		flag.StringVar(&fc.DSN, "d", "", "postgres DSN")
		flag.IntVar(&fc.StoreInterval, "i", defaultStoreIntervalSeconds, "store interval")
	}
	if fc.AppType == AppTypeAgent {
		flag.IntVar(&fc.PollInterval, "p", defaultPollIntervalSeconds, "poll interval")
		flag.IntVar(&fc.ReportInterval, "r", defaultReportIntervalSeconds, "report interval")
		flag.IntVar(&fc.RateLimit, "l", defaultRateLimit, "rate limit")
	}

	flag.StringVar(&fc.LogLevel, "log_level", "debug", "log level")
	flag.StringVar(&fc.Address, "a", "localhost:8080", "server address")
	flag.StringVar(&fc.HashKey, "k", "", "hash key")

	flag.Parse()

	logger := slog.Default()
	logger.Info("init flags",
		slog.String("LOG_LEVEL", fc.LogLevel),
		slog.String("ADDRESS", fc.Address),
		slog.String("KEY", fc.HashKey),
		slog.Int("REPORT_INTERVAL", fc.ReportInterval),
		slog.Int("POLL_INTERVAL", fc.PollInterval),
		slog.Int("RATE_LIMIT", fc.RateLimit),
		slog.Int("STORE_INTERVAL", fc.StoreInterval),
		slog.String("FILE_STORAGE_PATH", fc.FileStoragePath),
		slog.Bool("RESTORE", fc.Restore),
		slog.String("DATABASE_DSN", fc.DSN),
	)
}

func (ec *EnvContainer) Parse() {
	logger := slog.Default()
	logger.Info("init envs",
		slog.String("LOG_LEVEL", os.Getenv("LOG_LEVEL")),
		slog.String("ADDRESS", os.Getenv("ADDRESS")),
		slog.String("KEY", os.Getenv("KEY")),
		slog.String("REPORT_INTERVAL", os.Getenv("REPORT_INTERVAL")),
		slog.String("POLL_INTERVAL", os.Getenv("POLL_INTERVAL")),
		slog.String("RATE_LIMIT", os.Getenv("RATE_LIMIT")),
		slog.String("STORE_INTERVAL", os.Getenv("STORE_INTERVAL")),
		slog.String("FILE_STORAGE_PATH", os.Getenv("FILE_STORAGE_PATH")),
		slog.String("RESTORE", os.Getenv("RESTORE")),
		slog.String("DATABASE_DSN", os.Getenv("DATABASE_DSN")),
	)

	err := env.Parse(ec)
	if err != nil {
		logger.Error("fail to Parse env", slog.Any("error", err))
		panic(err)
	}
}

func buildBaseConfig(appType AppType, fc *FlagContainer, ec *EnvContainer) BaseConfig {
	conf := BaseConfig{
		AppType: appType,
	}

	if ec.Address != "" {
		conf.Address = ec.Address
	} else {
		conf.Address = fc.Address
	}

	if ec.LogLevel != "" {
		conf.LogLevel = ec.LogLevel
	} else {
		conf.LogLevel = fc.LogLevel
	}

	if ec.HashKey != "" {
		conf.HashKey = ec.HashKey
	} else {
		conf.HashKey = fc.HashKey
	}

	return conf
}

func NewAgentConfig(fc *FlagContainer, ec *EnvContainer) (*AgentConfig, error) {
	conf := &AgentConfig{BaseConfig: buildBaseConfig(AppTypeAgent, fc, ec)}

	if ec.ReportInterval != 0 {
		conf.ReportInterval = ec.ReportInterval
	} else {
		conf.ReportInterval = fc.ReportInterval
	}
	if ec.PollInterval != 0 {
		conf.PollInterval = ec.PollInterval
	} else {
		conf.PollInterval = fc.PollInterval
	}
	if ec.RateLimit != 0 {
		conf.RateLimit = ec.RateLimit
	} else {
		conf.RateLimit = fc.RateLimit
	}

	logger := slog.Default()
	logger.Info("final agent params",
		slog.String("ADDRESS", conf.Address),
		slog.Int("REPORT_INTERVAL", conf.ReportInterval),
		slog.Int("POLL_INTERVAL", conf.PollInterval),
		slog.Int("RATE_LIMIT", conf.RateLimit),
		slog.String("LOG_LEVEL", conf.LogLevel),
		slog.String("KEY", conf.HashKey),
	)

	return conf, nil
}

func NewServerConfig(fc *FlagContainer, ec *EnvContainer) (*ServerConfig, error) {
	conf := &ServerConfig{BaseConfig: buildBaseConfig(AppTypeServer, fc, ec)}

	if ec.FileStoragePath != "" {
		conf.FileStoragePath = ec.FileStoragePath
	} else {
		conf.FileStoragePath = fc.FileStoragePath
	}
	if ec.DSN != "" {
		conf.DSN = ec.DSN
	} else {
		conf.DSN = fc.DSN
	}

	v, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		vInt, vErr := strconv.Atoi(v)
		if vErr != nil {
			return nil, fmt.Errorf("convert STORE_INTERVAL env to int error: %w", vErr)
		}

		conf.StoreInterval = vInt
	} else {
		conf.StoreInterval = fc.StoreInterval
	}

	v, ok = os.LookupEnv("RESTORE")
	if ok {
		vBool, vBoolErr := strconv.ParseBool(v)
		if vBoolErr != nil {
			return nil, fmt.Errorf("convert RESTORE env to bool error: %w", vBoolErr)
		}

		conf.Restore = vBool
	} else {
		conf.Restore = fc.Restore
	}

	logger := slog.Default()
	logger.Info("final server params",
		slog.String("ADDRESS", conf.Address),
		slog.Bool("RESTORE", conf.Restore),
		slog.Int("STORE_INTERVAL", conf.StoreInterval),
		slog.String("FILE_STORAGE_PATH", conf.FileStoragePath),
		slog.String("LOG_LEVEL", conf.LogLevel),
		slog.String("KEY", conf.HashKey),
		slog.String("DATABASE_DSN", conf.DSN),
	)

	return conf, nil
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
