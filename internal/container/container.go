package container

import (
	"errors"
	"flag"
	"log"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap"
)

const (
	defaultReportIntervalSeconds = 10
	defaultPollIntervalSeconds   = 2
)

type (
	Container struct {
		Config *Config
		Logger *zap.Logger
	}
	Config struct {
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
	}
)

var (
	appContainer *Container
	once         sync.Once
)

func InitContainer() {
	once.Do(func() {
		config, err := initAppConfig()
		if err != nil {
			log.Panic(err)
		}

		logger := zap.Must(zap.NewDevelopment())

		defer func(logger *zap.Logger) {
			syncErr := logger.Sync()
			if syncErr != nil {
				if errors.Is(syncErr, syscall.EINVAL) {
					// Sync is not supported on os.Stderr / os.Stdout on all platforms.
					return
				}
				logger.Error("Failed to sync logger", zap.Error(syncErr))
			}
		}(logger)

		appContainer = &Container{Config: config, Logger: logger}

		logger.Sugar().Debug("init container success")
	})
}

func GetConfig() *Config {
	return appContainer.Config
}

func GetLogger() *zap.Logger {
	return appContainer.Logger
}

func initAppConfig() (*Config, error) {
	var (
		addr string
		ri   int
		pi   int
	)

	c := Config{}

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&addr, "a", "localhost:8080", "server host:port")
	flag.IntVar(&ri, "r", defaultReportIntervalSeconds, "report interval")
	flag.IntVar(&pi, "p", defaultPollIntervalSeconds, "poll interval")

	flag.Parse()

	if c.Address == "" {
		c.Address = addr
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = ri
	}
	if c.PollInterval == 0 {
		c.PollInterval = pi
	}

	return &c, nil
}

func (c Config) GetAddress() string {
	return c.Address
}

func (c Config) GetReportIntervalDuration() time.Duration {
	return time.Duration(c.ReportInterval) * time.Second
}

func (c Config) GetPollIntervalDuration() time.Duration {
	return time.Duration(c.PollInterval) * time.Second
}
