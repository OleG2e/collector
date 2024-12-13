package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerHostPort string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

var appConfig *Config

func GetConfig() *Config {
	if appConfig == nil {
		c, err := initAppConfig()
		if err != nil {
			log.Fatal(err)
		}
		appConfig = c
	}
	return appConfig
}

const defaultReportIntervalSeconds = 10
const defaultPollIntervalSeconds = 2

func initAppConfig() (*Config, error) {
	var (
		sh string
		ri int
		pi int
	)

	c := Config{}

	err := env.Parse(&c)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&sh, "a", "localhost:8080", "server host:port")
	flag.IntVar(&ri, "r", defaultReportIntervalSeconds, "report interval")
	flag.IntVar(&pi, "p", defaultPollIntervalSeconds, "poll interval")

	flag.Parse()

	if c.ServerHostPort == "" {
		c.ServerHostPort = sh
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = ri
	}
	if c.PollInterval == 0 {
		c.PollInterval = pi
	}

	return &c, nil
}

func (c Config) GetServerHostPort() string {
	return c.ServerHostPort
}

func (c Config) GetReportInterval() int {
	return c.ReportInterval
}

func (c Config) GetPollInterval() int {
	return c.PollInterval
}
