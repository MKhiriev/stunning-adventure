package config

import (
	"errors"
	"fmt"

	"dario.cat/mergo"
)

// ServerConfig
//
//	Configuration for the server component.
//	Controls server address, storage intervals, file persistence, database connection, hashing, and audit settings.
type ServerConfig struct {
	ServerAddress string `env:"ADDRESS"`        // network address to bind the server
	StoreInterval int64  `env:"STORE_INTERVAL"` // interval in seconds to persist metrics to storage

	FileStoragePath        string `env:"FILE_STORAGE_PATH"` // file path for metric storage
	RestoreMetricsFromFile bool   `env:"RESTORE"`           // flag to restore metrics from file on startup
	DatabaseDSN            string `env:"DATABASE_DSN"`      // database connection string
	HashKey                string `env:"KEY"`               // key used for hashing metric payloads
	AuditFile              string `env:"AUDIT_FILE"`        // local file path for audit logs
	AuditURL               string `env:"AUDIT_URL"`         // external URL to send audit logs
	PrivateCryptoKeyPath   string `env:"CRYPTO_KEY"`        // private key path for decryption of data from agent
	TrustedSubnet          string `env:"TRUSTED_SUBNET"`    // CIDR of the trusted subnet; accepts requests only from within this subnet

	JSONConfigFile string `env:"CONFIG"` // json file path with configs
}

type serverConfigBuilder struct {
	configs []*ServerConfig
	err     error
}

func newServerConfigBuilder() *serverConfigBuilder {
	return &serverConfigBuilder{
		configs: make([]*ServerConfig, 0, 4),
	}
}

func (b *serverConfigBuilder) build() (*ServerConfig, error) {
	if b.err != nil {
		return nil, fmt.Errorf("error occured during building config: %w", b.err)
	}

	serverConfig := new(ServerConfig)
	for _, cfg := range b.configs {
		if err := mergo.Merge(serverConfig, cfg); err != nil {
			return nil, fmt.Errorf("error merging configs: %w", err)
		}
	}

	return serverConfig, serverConfig.validate()
}

func (b *serverConfigBuilder) withDefaults() *serverConfigBuilder {
	defaultConfig := &ServerConfig{
		ServerAddress: defaultServerAddress,
		StoreInterval: defaultStoreInterval,
	}

	b.configs = append(b.configs, defaultConfig)
	return b
}

func (b *serverConfigBuilder) withJSON() *serverConfigBuilder {
	var jsonPath string
	isJSONSpecified := false

	for _, cfg := range b.configs {
		if cfg.JSONConfigFile != "" {
			isJSONSpecified = true
			jsonPath = cfg.JSONConfigFile
		}
	}

	if isJSONSpecified {
		jsonCfg, err := parseServerJSON(jsonPath)
		if err != nil {
			b.err = errors.Join(b.err, err)
			return b
		}
		b.configs = append(b.configs, jsonCfg)
	}

	return b
}

func (b *serverConfigBuilder) withFlags() *serverConfigBuilder {
	flags := ParseServerFlags()

	b.configs = append(b.configs, flags)
	return b
}

func (b *serverConfigBuilder) withEnv() *serverConfigBuilder {
	envCfg := &ServerConfig{}
	if err := parseEnv(envCfg); err != nil {
		b.err = errors.Join(b.err, err)
		return b
	}

	b.configs = append(b.configs, envCfg)

	return b
}

// validate
//
// Description:
//
//	Validates essential fields of ServerConfig.
//	Ensures ServerAddress and StoreInterval are set.
//
// Returns:
//
//	error - descriptive error if validation fails, nil if valid
func (s *ServerConfig) validate() error {
	switch {
	case s.ServerAddress == "":
		return errors.New("invalid Server Address")
	case s.StoreInterval == 0:
		return errors.New("invalid Store Interval")
	}

	return nil
}

func (s *ServerConfig) setDefault() {
	if s.ServerAddress == "" {
		s.ServerAddress = defaultServerAddress
	}
	if s.StoreInterval == 0 {
		s.StoreInterval = defaultStoreInterval
	}
}

func (s *ServerConfig) isEmpty() bool {
	if s.ServerAddress != "" && s.StoreInterval != 0 {
		return false
	}

	return true
}
