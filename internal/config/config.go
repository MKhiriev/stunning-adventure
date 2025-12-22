package config

const (
	defaultPollInterval   = int64(2)
	defaultReportInterval = int64(5)
	defaultServerAddress  = "localhost:8080"
	defaultRateLimit      = int64(1)
	defaultStoreInterval  = int64(300)
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
	cfg := new(ServerConfig)

	// 1. Get ENV configs
	err := parseEnv(cfg)
	if err != nil {
		return nil, err
	}

	// if ENV configs are not empty
	if !cfg.isEmpty() {
		return cfg, cfg.validate()
	}

	// 2. Get CMD line configs
	cmdCfg := ParseServerFlags()

	fillEmptyServerConfigParams(cfg, cmdCfg)

	// if CMD configs are not empty
	if !cfg.isEmpty() {
		return cfg, nil
	}

	// 3. Get JSON configs
	if cfg.JSONConfigFile != "" {
		jsonCfg, err := parseServerJSON(cfg.JSONConfigFile)
		if err != nil {
			return nil, err
		}

		fillEmptyServerConfigParams(cfg, jsonCfg)
	}

	// 4. Set defaults if empty
	if cfg.isEmpty() {
		cfg.setDefault()
	}

	return cfg, nil
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

func fillEmptyServerConfigParams(to, from *ServerConfig) {
	if to.ServerAddress == "" {
		to.ServerAddress = from.ServerAddress
	}
	if to.StoreInterval == 0 {
		to.StoreInterval = from.StoreInterval
	}
	if to.FileStoragePath == "" {
		to.FileStoragePath = from.FileStoragePath
	}
	if !to.RestoreMetricsFromFile {
		to.RestoreMetricsFromFile = from.RestoreMetricsFromFile
	}
	if to.DatabaseDSN == "" {
		to.DatabaseDSN = from.DatabaseDSN
	}
	if to.HashKey == "" {
		to.HashKey = from.HashKey
	}
	if to.AuditFile == "" {
		to.AuditFile = from.AuditFile
	}
	if to.AuditURL == "" {
		to.AuditURL = from.AuditURL
	}
	if to.PrivateCryptoKeyPath == "" {
		to.PrivateCryptoKeyPath = from.PrivateCryptoKeyPath
	}
	if to.JSONConfigFile == "" {
		to.JSONConfigFile = from.JSONConfigFile
	}
}
