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

type serverJSONConfig struct {
	ServerAddress          string         `json:"address"`        // network address to bind the server
	GrpcServerAddress      string         `json:"grpc_address"`   // network address to bind the GRPC server
	StoreInterval          utils.Duration `json:"store_interval"` // interval in seconds to persist metrics to storage
	FileStoragePath        string         `json:"store_file"`     // file path for metric storage
	RestoreMetricsFromFile bool           `json:"restore"`        // flag to restore metrics from file on startup
	DatabaseDSN            string         `json:"database_dsn"`   // database connection string
	HashKey                string         `json:"hash_key"`       // key used for hashing metric payloads
	AuditFile              string         `json:"audit_file"`     // local file path for audit logs
	AuditURL               string         `json:"audit_url"`      // external URL to send audit logs
	PrivateCryptoKeyPath   string         `json:"crypto_key"`     // private key path for decryption of data from agent
	TrustedSubnet          string         `json:"trusted_subnet"` // CIDR of the trusted subnet; accepts requests only from within this subnet
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

func parseServerJSON(jsonFilePath string) (*ServerConfig, error) {
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading a json file: %w", err)
	}

	var jsonCfg serverJSONConfig
	if err := json.NewDecoder(jsonFile).Decode(&jsonCfg); err != nil {
		return nil, fmt.Errorf("error decoding json configs: %w", err)
	}

	return &ServerConfig{
		ServerAddress:          jsonCfg.ServerAddress,
		GrpcServerAddress:      jsonCfg.GrpcServerAddress,
		StoreInterval:          int64(jsonCfg.StoreInterval.Seconds()),
		FileStoragePath:        jsonCfg.FileStoragePath,
		RestoreMetricsFromFile: jsonCfg.RestoreMetricsFromFile,
		DatabaseDSN:            jsonCfg.DatabaseDSN,
		HashKey:                jsonCfg.HashKey,
		AuditFile:              jsonCfg.AuditFile,
		AuditURL:               jsonCfg.AuditURL,
		PrivateCryptoKeyPath:   jsonCfg.PrivateCryptoKeyPath,
		TrustedSubnet:          jsonCfg.TrustedSubnet,
	}, nil
}
