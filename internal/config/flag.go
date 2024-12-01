package config

import (
	"flag"
)

type Config struct {
	serverHostPort string
	reportInterval int
	pollInterval   int
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

	flag.StringVar(&sh, "a", "localhost:8080", "server host:port")
	flag.IntVar(&ri, "r", 10, "report interval")
	flag.IntVar(&pi, "p", 2, "poll interval")

	flag.Parse()

	c.pollInterval = pi
	c.reportInterval = ri
	c.serverHostPort = sh

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
