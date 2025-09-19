package config

import (
	"errors"
	"log"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	HashKey        string `env:"KEY"`
}

type ServerConfig struct {
	ServerAddress          string `env:"ADDRESS"`
	StoreInterval          int64  `env:"STORE_INTERVAL"`
	FileStoragePath        string `env:"FILE_STORAGE_PATH"`
	RestoreMetricsFromFile bool   `env:"RESTORE"`
	DatabaseDSN            string `env:"DATABASE_DSN"`
	HashKey                string `env:"KEY"`
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
	commandLineServerAddress, commandLinePollInterval, commandLineReportInterval, commandLineHashKey := ParseAgentFlags()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = commandLineServerAddress
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = commandLinePollInterval
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = commandLineReportInterval
	}
	if cfg.HashKey == "" {
		cfg.HashKey = commandLineHashKey
	}

	return cfg
}

func GetServerConfigs() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// if all values are not nil return cfg
	if cfg.ServerAddress != "" && cfg.StoreInterval != 0 {
		return cfg, cfg.Validate()
	}

	// else get command line args or default values
	commandLineServerAddress, commandLineStoreInterval, commandLineFileStoragePath, commandLineRestore, databaseDSN, commandLineHashKey := ParseServerFlags()

	if cfg.ServerAddress == "" {
		cfg.ServerAddress = commandLineServerAddress
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = commandLineStoreInterval
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = commandLineFileStoragePath
	}
	if !cfg.RestoreMetricsFromFile {
		cfg.RestoreMetricsFromFile = commandLineRestore
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = databaseDSN
	}
	if cfg.HashKey == "" {
		cfg.HashKey = commandLineHashKey
	}

	return cfg, cfg.Validate()
}

func (s *ServerConfig) Validate() error {
	switch {
	case s.ServerAddress == "":
		return errors.New("invalid Server Address")
	case s.StoreInterval == 0:
		return errors.New("invalid Store Interval")
	}

	return nil
}
