package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	serverHostPort string `env:"ADDRESS"`
	reportInterval int    `env:"REPORT_INTERVAL"`
	pollInterval   int    `env:"POLL_INTERVAL"`
}

var appConfig *Config

func GetConfig() *Config {
	if appConfig == nil {
		appConfig = initAppConfig()
	}
	return appConfig
}

func initAppConfig() *Config {
	var (
		sh string
		ri int
		pi int
	)

	c := Config{}

	err := env.Parse(&c)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&sh, "a", "localhost:8080", "server host:port")
	flag.IntVar(&ri, "r", 10, "report interval")
	flag.IntVar(&pi, "p", 2, "poll interval")

	flag.Parse()

	if c.serverHostPort == "" {
		c.serverHostPort = sh
	}
	if c.reportInterval == 0 {
		c.reportInterval = ri
	}
	if c.pollInterval == 0 {
		c.pollInterval = pi
	}

	return &c
}

func (c Config) GetServerHostPort() string {
	return c.serverHostPort
}

func (c Config) GetReportInterval() int {
	return c.reportInterval
}

func (c Config) GetPollInterval() int {
	return c.pollInterval
}
