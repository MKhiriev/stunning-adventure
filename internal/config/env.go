package config

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
}

type ServerConfig struct {
	ServerAddress string `env:"ADDRESS"` // can define if anything was passed
}

func GetAgentConfigs() *AgentConfig {
	cfg := &AgentConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// if all values are not nil return cfg
	if cfg.ServerAddress != "" && cfg.ReportInterval != 0 && cfg.PollInterval != 0 {
		return cfg
	}

	// else get command line args or default values
	commandLineServerAddress, commandLinePollInterval, commandLineReportInterval := ParseAgentFlags()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = commandLineServerAddress
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = commandLinePollInterval
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = commandLineReportInterval
	}

	return cfg
}

func GetServerConfigs() *ServerConfig {
	cfg := &ServerConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// if serverAddress is not nil return cfg
	if cfg.ServerAddress != "" {
		return cfg
	}

	// else get command line args or default value
	serverAddress := ParseServerFlags()
	cfg.ServerAddress = serverAddress

	return cfg
}
