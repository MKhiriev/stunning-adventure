package config

import (
	"errors"
	"fmt"

	"dario.cat/mergo"
)

// AgentConfig
//
//	Configuration for the agent component of the system.
//	Holds network, reporting, polling, hashing, and rate-limiting settings.
type AgentConfig struct {
	ServerAddress  string `env:"ADDRESS"`          // address of the server to send metrics to
	ReportInterval int64  `env:"REPORT_INTERVAL" ` // interval in seconds between sending metrics to server
	PollInterval   int64  `env:"POLL_INTERVAL" `   // interval in seconds to read metrics from system
	RateLimit      int64  `env:"RATE_LIMIT" `      // maximum number of requests per second

	HashKey             string `env:"KEY"`         // key used for hashing metric payloads
	PublicCryptoKeyPath string `env:"CRYPTO_KEY" ` // public key path for encryption of data for server
	JSONConfigFile      string `env:"CONFIG"`      // json file path with configs
}

type agentConfigBuilder struct {
	configs []*AgentConfig
	err     error
}

func newAgentConfigBuilder() *agentConfigBuilder {
	return &agentConfigBuilder{
		configs: make([]*AgentConfig, 0, 4),
	}
}

func (b *agentConfigBuilder) build() (*AgentConfig, error) {
	if b.err != nil {
		return nil, fmt.Errorf("error occured during building config: %w", b.err)
	}

	agentConfig := new(AgentConfig)
	for _, cfg := range b.configs {
		if err := mergo.Merge(agentConfig, cfg); err != nil {
			return nil, fmt.Errorf("error merging configs: %w", err)
		}
	}

	return agentConfig, nil
}

func (b *agentConfigBuilder) withDefaults() *agentConfigBuilder {
	defaultConfig := &AgentConfig{
		ServerAddress:  defaultServerAddress,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		RateLimit:      defaultRateLimit,
	}

	b.configs = append(b.configs, defaultConfig)
	return b
}

func (b *agentConfigBuilder) withJSON() *agentConfigBuilder {
	var jsonPath string
	isJSONSpecified := false

	for _, cfg := range b.configs {
		if cfg.JSONConfigFile != "" {
			isJSONSpecified = true
			jsonPath = cfg.JSONConfigFile
		}
	}

	if isJSONSpecified {
		jsonCfg, err := parseAgentJSON(jsonPath)
		if err != nil {
			b.err = errors.Join(b.err, err)
			return b
		}
		b.configs = append(b.configs, jsonCfg)
	}

	return b
}

func (b *agentConfigBuilder) withFlags() *agentConfigBuilder {
	flags := parseAgentFlags()

	b.configs = append(b.configs, flags)
	return b
}

func (b *agentConfigBuilder) withEnv() *agentConfigBuilder {
	envCfg := &AgentConfig{}
	if err := parseEnv(envCfg); err != nil {
		b.err = errors.Join(b.err, err)
		return b
	}

	b.configs = append(b.configs, envCfg)
	return b
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
