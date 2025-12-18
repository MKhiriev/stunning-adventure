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

func fillEmptyServerConfigParams(cfg, from *ServerConfig) {
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = from.ServerAddress
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = from.StoreInterval
	}
	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = from.FileStoragePath
	}
	if !cfg.RestoreMetricsFromFile {
		cfg.RestoreMetricsFromFile = from.RestoreMetricsFromFile
	}
	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = from.DatabaseDSN
	}
	if cfg.HashKey == "" {
		cfg.HashKey = from.HashKey
	}
	if cfg.AuditFile == "" {
		cfg.AuditFile = from.AuditFile
	}
	if cfg.AuditURL == "" {
		cfg.AuditURL = from.AuditURL
	}
	if cfg.PrivateCryptoKeyPath == "" {
		cfg.PrivateCryptoKeyPath = from.PrivateCryptoKeyPath
	}
}
