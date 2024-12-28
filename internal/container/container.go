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
	defaultStoreIntervalSeconds  = 300
)

type (
	AgentContainer struct {
		Config *AgentConfig
	}
	ServerContainer struct {
		Config *ServerConfig
	}
	AgentConfig struct {
		Address        string `env:"ADDRESS"`
		ReportInterval int    `env:"REPORT_INTERVAL"`
		PollInterval   int    `env:"POLL_INTERVAL"`
	}
	ServerConfig struct {
		Address         string `env:"ADDRESS"`
		StoreInterval   int    `env:"STORE_INTERVAL"`
		FileStoragePath string `env:"FILE_STORAGE_PATH"`
		Restore         bool   `env:"RESTORE"`
	}
)

var (
	appAgentContainer  *AgentContainer
	appServerContainer *ServerContainer
	logger             *zap.SugaredLogger
	once               sync.Once
)

func InitAgentContainer() {
	once.Do(func() {
		config, err := initAppAgentConfig()
		if err != nil {
			log.Panic(err)
		}

		appAgentContainer = &AgentContainer{Config: config}

		initLogger()

		GetLogger().Debug("init agent container success")
	})
}

func InitServerContainer() {
	once.Do(func() {
		config, err := initAppServerConfig()
		if err != nil {
			log.Panic(err)
		}

		appServerContainer = &ServerContainer{Config: config}

		initLogger()

		GetLogger().Debug("init server container success")
	})
}

func initLogger() {
	logger = zap.Must(zap.NewDevelopment()).Sugar()
	defer func(logger *zap.SugaredLogger) {
		syncErr := logger.Sync()
		if syncErr != nil {
			if errors.Is(syncErr, syscall.EINVAL) {
				// Sync is not supported on os.Stderr / os.Stdout on all platforms.
				return
			}
			logger.Error("Failed to sync logger", zap.Error(syncErr))
		}
	}(logger)
}

func GetAgentConfig() *AgentConfig {
	return appAgentContainer.Config
}

func GetServerConfig() *ServerConfig {
	return appServerContainer.Config
}

func GetLogger() *zap.SugaredLogger {
	return logger
}

func initAppAgentConfig() (*AgentConfig, error) {
	var (
		addr string
		ri   int
		pi   int
	)

	c := AgentConfig{}

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

func initAppServerConfig() (*ServerConfig, error) {
	var (
		addr string
		fs   string
		si   int
		r    bool
	)

	c := ServerConfig{}

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&addr, "a", "localhost:8080", "server host:port")
	flag.IntVar(&si, "i", defaultStoreIntervalSeconds, "store interval")
	flag.StringVar(&fs, "f", ".", "file storage path")
	flag.BoolVar(&r, "r", true, "restore previous data")

	flag.Parse()

	if c.Address == "" {
		c.Address = addr
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = fs
	}

	if c.StoreInterval == 0 {
		c.StoreInterval = si
	}
	if !c.Restore {
		c.Restore = r
	}

	return &c, nil
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
