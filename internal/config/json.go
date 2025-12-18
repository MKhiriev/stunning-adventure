package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

type agentJSONConfig struct {
	ServerAddress  string         `json:"address"`         // address of the server to send metrics to
	ReportInterval utils.Duration `json:"report_interval"` // interval in seconds between sending metrics to server
	PollInterval   utils.Duration `json:"poll_interval"`   // interval in seconds to read metrics from system
	RateLimit      int64          `json:"rate_limit"`      // maximum number of requests per second

	HashKey             string `json:"hash_key"`   // key used for hashing metric payloads
	PublicCryptoKeyPath string `json:"crypto_key"` // public key path for encryption of data for server
}

func parseAgentJSON(jsonFilePath string) (*AgentConfig, error) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading a json file: %w", err)
	}

	var jsonCfg agentJSONConfig
	if err := json.NewDecoder(jsonFile).Decode(&jsonCfg); err != nil {
		return nil, fmt.Errorf("error decoding json configs: %w", err)
	}

	return &AgentConfig{
		ServerAddress:       jsonCfg.ServerAddress,
		ReportInterval:      int64(jsonCfg.ReportInterval.Seconds()),
		PollInterval:        int64(jsonCfg.PollInterval.Seconds()),
		RateLimit:           jsonCfg.RateLimit,
		HashKey:             jsonCfg.HashKey,
		PublicCryptoKeyPath: jsonCfg.PublicCryptoKeyPath,
	}, nil
}
