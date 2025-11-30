// Package config provides flag-parsing utilities for server and agent binaries.
// It defines configuration defaults, a NetAddress type implementing flag.Value,
// and parsing helpers returning fully resolved configuration parameters.
package config

import (
	"errors"
	"log"

	"github.com/caarlos0/env/v11"
)

// AgentConfig
//
//	Configuration for the agent component of the system.
//	Holds network, reporting, polling, hashing, and rate-limiting settings.
type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS"`         // address of the server to send metrics to
	ReportInterval int64  `env:"REPORT_INTERVAL"` // interval in seconds between sending metrics to server
	PollInterval   int64  `env:"POLL_INTERVAL"`   // interval in seconds to read metrics from system
	HashKey        string `env:"KEY"`             // key used for hashing metric payloads
	RateLimit      int64  `env:"RATE_LIMIT"`      // maximum number of requests per second
}

// ServerConfig
//
//	Configuration for the server component.
//	Controls server address, storage intervals, file persistence, database connection, hashing, and audit settings.
type ServerConfig struct {
	ServerAddress          string `env:"ADDRESS"`           // network address to bind the server
	StoreInterval          int64  `env:"STORE_INTERVAL"`    // interval in seconds to persist metrics to storage
	FileStoragePath        string `env:"FILE_STORAGE_PATH"` // file path for metric storage
	RestoreMetricsFromFile bool   `env:"RESTORE"`           // flag to restore metrics from file on startup
	DatabaseDSN            string `env:"DATABASE_DSN"`      // database connection string
	HashKey                string `env:"KEY"`               // key used for hashing metric payloads

	AuditFile string `env:"AUDIT_FILE"` // local file path for audit logs
	AuditURL  string `env:"AUDIT_URL"`  // external URL to send audit logs
}

// GetAgentConfigs
//
// Description:
//
//	Loads agent configuration from environment variables.
//	Falls back to command line flags or defaults if env variables are missing.
//
// Returns:
//
//	*AgentConfig - fully populated agent configuration struct
func GetAgentConfigs() *AgentConfig {
	cfg := &AgentConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// if all values are not nil return cfg
	if cfg.ServerAddress != "" && cfg.ReportInterval != 0 && cfg.PollInterval != 0 && cfg.RateLimit != 0 {
		return cfg
	}

	// else get command line args or default values
	commandLineServerAddress, commandLinePollInterval, commandLineReportInterval, commandLineHashKey, commandLineRateLimit := ParseAgentFlags()

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
	if cfg.RateLimit == 0 {
		cfg.RateLimit = commandLineRateLimit
	}

	return cfg
}

// GetServerConfigs
//
// Description:
//
//	Loads server configuration from environment variables.
//	Falls back to command line flags or defaults if env variables are missing.
//	Validates required fields before returning.
//
// Returns:
//
//	*ServerConfig - fully populated server configuration struct
//	error         - validation error if required fields are missing or invalid
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
	commandLineServerAddress, commandLineStoreInterval, commandLineFileStoragePath,
		commandLineRestore, databaseDSN, commandLineHashKey,
		commandLineAuditFilePath, commandLineAuditURL := ParseServerFlags()

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
	if cfg.AuditFile == "" {
		cfg.AuditFile = commandLineAuditFilePath
	}
	if cfg.AuditURL == "" {
		cfg.AuditURL = commandLineAuditURL
	}

	return cfg, cfg.Validate()
}

// Validate
//
// Description:
//
//	Validates essential fields of ServerConfig.
//	Ensures ServerAddress and StoreInterval are set.
//
// Returns:
//
//	error - descriptive error if validation fails, nil if valid
func (s *ServerConfig) Validate() error {
	switch {
	case s.ServerAddress == "":
		return errors.New("invalid Server Address")
	case s.StoreInterval == 0:
		return errors.New("invalid Store Interval")
	}

	return nil
}
