package config

import (
	"github.com/caarlos0/env/v11"
)

const (
	defaultPollInterval       = int64(2)
	defaultReportInterval     = int64(5)
	defaultServerAddress      = "localhost:8080"
	defaultRateLimit          = int64(1)
	defaultStoreInterval      = int64(300)
	defaultFileStoragePath    = ""
	defaultRestoreValue       = false
	defaultDatabaseDSN        = ""
	defaultHashKey            = ""
	defaultAuditFilePath      = ""
	defaultAuditURL           = ""
	defaultPublicKeyFilePath  = ""
	defaultPrivateKeyFilePath = ""
	defaultJSONConfigFilePath = ""
)

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
func GetAgentConfigs() (*AgentConfig, error) {
	cfg := new(AgentConfig)

	// 1. Get ENV configs
	err := parseEnv(cfg)
	if err != nil {
		return nil, err
	}

	// if ENV configs are not empty
	if !cfg.isEmpty() {
		return cfg, nil
	}

	// 2. Get CMD line configs
	cmdCfg := parseAgentFlags()

	fillEmptyAgentConfigParams(cfg, cmdCfg)

	// if CMD configs are not empty
	if !cfg.isEmpty() {
		return cfg, nil
	}

	// 3. Get JSON configs
	if cfg.JSONConfigFile != "" {
		jsonCfg, err := parseAgentJSON(cfg.JSONConfigFile)
		if err != nil {
			return nil, err
		}

		fillEmptyAgentConfigParams(cfg, jsonCfg)
	}

	// 4. Set defaults if empty
	if cfg.isEmpty() {
		cfg.setDefault()
	}

	return cfg, nil
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
		return nil, err
	}

	// if all values are not nil return cfg
	if cfg.ServerAddress != "" && cfg.StoreInterval != 0 {
		return cfg, cfg.validate()
	}

	// else get command line args or default values
	commandLineServerAddress, commandLineStoreInterval, commandLineFileStoragePath,
		commandLineRestore, databaseDSN, commandLineHashKey,
		commandLineAuditFilePath, commandLineAuditURL, commandLinePrivateKeyFilePath := ParseServerFlags()

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
	if cfg.PrivateCryptoKeyPath == "" {
		cfg.PrivateCryptoKeyPath = commandLinePrivateKeyFilePath
	}

	return cfg, cfg.validate()
}

func fillEmptyAgentConfigParams(to, from *AgentConfig) {
	if to.ServerAddress == "" {
		to.ServerAddress = from.ServerAddress
	}
	if to.PollInterval == 0 {
		to.PollInterval = from.PollInterval
	}
	if to.ReportInterval == 0 {
		to.ReportInterval = from.ReportInterval
	}
	if to.HashKey == "" {
		to.HashKey = from.HashKey
	}
	if to.RateLimit == 0 {
		to.RateLimit = from.RateLimit
	}
	if to.PublicCryptoKeyPath == "" {
		to.PublicCryptoKeyPath = from.PublicCryptoKeyPath
	}
	if to.JSONConfigFile == "" {
		to.JSONConfigFile = from.JSONConfigFile
	}
}
