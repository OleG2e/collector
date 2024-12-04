package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	ServerHostPort string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
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

	if c.ServerHostPort == "" {
		c.ServerHostPort = sh
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = ri
	}
	if c.PollInterval == 0 {
		c.PollInterval = pi
	}

	return &c
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
