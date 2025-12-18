package config

// AgentConfig
//
//	Configuration for the agent component of the system.
//	Holds network, reporting, polling, hashing, and rate-limiting settings.
type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS" json:"address"`                 // address of the server to send metrics to
	ReportInterval int64  `env:"REPORT_INTERVAL" json:"report_interval"` // interval in seconds between sending metrics to server
	PollInterval   int64  `env:"POLL_INTERVAL" json:"poll_interval"`     // interval in seconds to read metrics from system
	RateLimit      int64  `env:"RATE_LIMIT" json:"rate_limit"`           // maximum number of requests per second

	HashKey             string `env:"KEY" json:"hash_key"`          // key used for hashing metric payloads
	PublicCryptoKeyPath string `env:"CRYPTO_KEY" json:"crypto_key"` // public key path for encryption of data for server
	JSONConfigFile      string `env:"CONFIG" json:"-"`              // json file path with configs
}

func (a *AgentConfig) setDefault() {
	if a.ServerAddress == "" {
		a.ServerAddress = defaultServerAddress
	}
	if a.ReportInterval == 0 {
		a.ReportInterval = defaultReportInterval
	}
	if a.PollInterval == 0 {
		a.PollInterval = defaultPollInterval
	}
	if a.RateLimit == 0 {
		a.RateLimit = defaultRateLimit
	}
}

func (a *AgentConfig) isEmpty() bool {
	if a.ServerAddress != "" && a.ReportInterval != 0 && a.PollInterval != 0 && a.RateLimit != 0 {
		return false
	}

	return true
}
