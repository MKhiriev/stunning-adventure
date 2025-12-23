package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

func TestParseAgentJSON_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "agent.json")

	// создаём JSON с правильными данными
	agentCfg := agentJSONConfig{
		ServerAddress:       "127.0.0.1:8080",
		ReportInterval:      utils.Duration{Duration: time.Second * 5},
		PollInterval:        utils.Duration{Duration: time.Second * 2},
		RateLimit:           100,
		HashKey:             "abc",
		PublicCryptoKeyPath: "pub.key",
	}
	data, _ := json.Marshal(agentCfg)
	if err := os.WriteFile(filePath, data, 0777); err != nil {
		t.Fatalf("failed to write JSON file: %v", err)
	}

	cfg, err := parseAgentJSON(filePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ServerAddress != agentCfg.ServerAddress {
		t.Errorf("unexpected ServerAddress: got %s", cfg.ServerAddress)
	}
	if cfg.ReportInterval != int64(time.Second*5/time.Second) {
		t.Errorf("unexpected ReportInterval: got %d", cfg.ReportInterval)
	}
	if cfg.HashKey != agentCfg.HashKey {
		t.Errorf("unexpected HashKey: got %s", cfg.HashKey)
	}
	if cfg.PublicCryptoKeyPath != agentCfg.PublicCryptoKeyPath {
		t.Errorf("unexpected PublicCryptoKeyPath: got %s", cfg.PublicCryptoKeyPath)
	}
}

func TestParseAgentJSON_FileNotExist(t *testing.T) {
	_, err := parseAgentJSON("nonexistent.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestParseAgentJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "agent.json")
	_ = os.WriteFile(filePath, []byte("{invalid json}"), 0644)

	_, err := parseAgentJSON(filePath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseServerJSON_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "server.json")

	serverCfg := serverJSONConfig{
		ServerAddress:          "127.0.0.1:9090",
		StoreInterval:          utils.Duration{Duration: time.Second * 10},
		FileStoragePath:        "/tmp/store",
		RestoreMetricsFromFile: true,
		DatabaseDSN:            "postgres://user:pass@localhost/db",
		HashKey:                "hash",
		AuditFile:              "audit.log",
		AuditURL:               "http://audit",
		PrivateCryptoKeyPath:   "priv.key",
	}
	data, _ := json.Marshal(serverCfg)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("failed to write JSON file: %v", err)
	}

	cfg, err := parseServerJSON(filePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ServerAddress != serverCfg.ServerAddress {
		t.Errorf("unexpected ServerAddress: got %s", cfg.ServerAddress)
	}
	if cfg.StoreInterval != int64(time.Second*10/time.Second) {
		t.Errorf("unexpected StoreInterval: got %d", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != serverCfg.FileStoragePath {
		t.Errorf("unexpected FileStoragePath: got %s", cfg.FileStoragePath)
	}
	if cfg.RestoreMetricsFromFile != serverCfg.RestoreMetricsFromFile {
		t.Errorf("unexpected RestoreMetricsFromFile: got %v", cfg.RestoreMetricsFromFile)
	}
	if cfg.DatabaseDSN != serverCfg.DatabaseDSN {
		t.Errorf("unexpected DatabaseDSN: got %s", cfg.DatabaseDSN)
	}
}

func TestParseServerJSON_FileNotExist(t *testing.T) {
	_, err := parseServerJSON("nonexistent.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestParseServerJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "server.json")
	_ = os.WriteFile(filePath, []byte("{invalid json}"), 0644)

	_, err := parseServerJSON(filePath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
