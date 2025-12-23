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
	return newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
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
	return newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
}
