package config

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/OleG2e/collector/pkg/logging"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

const (
	defaultReportIntervalSeconds = 10
	defaultPollIntervalSeconds   = 2
	defaultStoreIntervalSeconds  = 300
)

type (
	AgentConfig struct {
		LogLevel       string `env:"LOG_LEVEL"`
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
	}
	ServerConfig struct {
		LogLevel        string `env:"LOG_LEVEL"`
		Address         string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
		DSN             string `env:"DATABASE_DSN"`
	}
)

func NewAgentConfig(ctx context.Context, l *logging.ZapLogger) (*AgentConfig, error) {
	var (
		addr     string
		logLevel string
		ri       int
		pi       int
	)

	c := AgentConfig{}

	err := env.Parse(&c)

	l.DebugCtx(ctx, "init agent env config",
		zap.String("LOG_LEVEL", os.Getenv("LOG_LEVEL")),
		zap.String("ADDRESS", os.Getenv("ADDRESS")),
		zap.String("REPORT_INTERVAL", os.Getenv("REPORT_INTERVAL")),
		zap.String("POLL_INTERVAL", os.Getenv("POLL_INTERVAL")),
	)

	if err != nil {
		return nil, err
	}

	flag.StringVar(&logLevel, "log_level", "info", "log level")
	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.IntVar(&ri, "r", defaultReportIntervalSeconds, "report interval")
	flag.IntVar(&pi, "p", defaultPollIntervalSeconds, "poll interval")

	flag.Parse()

	l.DebugCtx(ctx, "init agent flags config",
		zap.String("LOG_LEVEL", logLevel),
		zap.String("ADDRESS", addr),
		zap.Int("REPORT_INTERVAL", ri),
		zap.Int("POLL_INTERVAL", pi),
	)

	if c.LogLevel == "" {
		c.LogLevel = logLevel
	}
	if c.Address == "" {
		c.Address = addr
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = ri
	}
	if c.PollInterval == 0 {
		c.PollInterval = pi
	}

	l.DebugCtx(ctx, "final agent params",
		zap.String("ADDRESS", c.Address),
		zap.Int("REPORT_INTERVAL", c.ReportInterval),
		zap.Int("POLL_INTERVAL", c.PollInterval),
		zap.String("LOG_LEVEL", c.LogLevel),
	)

	return &c, nil
}

func NewServerConfig(ctx context.Context, l *logging.ZapLogger) (*ServerConfig, error) {
	var (
		addr     string
		logLevel string
		fs       string
		dsn      string
		si       int
		r        bool
	)

	c := ServerConfig{}

	err := env.Parse(&c)

	l.DebugCtx(ctx, "init server env config",
		zap.String("ADDRESS", os.Getenv("ADDRESS")),
		zap.String("LOG_LEVEL", logLevel),
		zap.String("STORE_INTERVAL", os.Getenv("STORE_INTERVAL")),
		zap.String("FILE_STORAGE_PATH", os.Getenv("FILE_STORAGE_PATH")),
		zap.String("RESTORE", os.Getenv("RESTORE")),
		zap.String("RESTORE", os.Getenv("DATABASE_DSN")),
	)

	if err != nil {
		return nil, err
	}

	flag.StringVar(&addr, "a", "localhost:8080", "server host:port")
	flag.StringVar(&logLevel, "log_level", "info", "log level")
	flag.IntVar(&si, "i", defaultStoreIntervalSeconds, "store interval")
	flag.StringVar(&fs, "f", "storage.db", "file storage path")
	flag.BoolVar(&r, "r", true, "restore previous data")
	flag.StringVar(&dsn, "d", "", "postgres dsn")

	flag.Parse()

	l.DebugCtx(ctx, "init server flags config",
		zap.String("ADDRESS", addr),
		zap.String("FILE_STORAGE_PATH", fs),
		zap.String("LOG_LEVEL", logLevel),
		zap.String("DATABASE_DSN", dsn),
		zap.Int("STORE_INTERVAL", si),
		zap.Bool("RESTORE", r),
	)

	if c.LogLevel == "" {
		c.LogLevel = logLevel
	}
	if c.Address == "" {
		c.Address = addr
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = fs
	}
	if c.DSN == "" {
		c.DSN = dsn
	}

	v, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		vInt, vErr := strconv.Atoi(v)
		if vErr != nil {
			return nil, vErr
		}
		c.StoreInterval = vInt
	} else {
		c.StoreInterval = si
	}

	v, ok = os.LookupEnv("RESTORE")
	if ok {
		vBool, vBoolErr := strconv.ParseBool(v)
		if vBoolErr != nil {
			return nil, vBoolErr
		}
		c.Restore = vBool
	} else {
		c.Restore = r
	}

	l.DebugCtx(ctx, "final server params",
		zap.String("ADDRESS", c.Address),
		zap.String("FILE_STORAGE_PATH", c.FileStoragePath),
		zap.Int("STORE_INTERVAL", c.StoreInterval),
		zap.Bool("RESTORE", c.Restore),
		zap.String("LOG_LEVEL", c.LogLevel),
		zap.String("DATABASE_DSN", c.DSN),
	)

	return &c, nil
}

func (c *AgentConfig) GetLogLevel() zapcore.Level {
	level, levelErr := zap.ParseAtomicLevel(c.LogLevel)
	if levelErr != nil {
		log.Panic(levelErr)
	}

	return level.Level()
}

func (c *ServerConfig) GetLogLevel() zapcore.Level {
	level, levelErr := zap.ParseAtomicLevel(c.LogLevel)
	if levelErr != nil {
		log.Panic(levelErr)
	}

	return level.Level()
}

func (c *AgentConfig) GetAddress() string {
	return c.Address
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
