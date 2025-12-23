package config

import (
	"os"
	"testing"
)

func TestParseEnv_AgentConfig(t *testing.T) {
	// временно сохраняем и восстанавливаем переменные окружения
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range origEnv {
			parts := splitEnv(kv)
			os.Setenv(parts[0], parts[1])
		}
	}()

	// устанавливаем тестовые переменные окружения
	os.Setenv("ADDRESS", "127.0.0.1:8080")
	os.Setenv("REPORT_INTERVAL", "10")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("RATE_LIMIT", "42")
	os.Setenv("KEY", "hashkey")
	os.Setenv("CRYPTO_KEY", "/tmp/pub.key")
	os.Setenv("CONFIG", "/tmp/config.json")

	cfg := &AgentConfig{}
	if err := parseEnv(cfg); err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:8080" {
		t.Errorf("unexpected ServerAddress: %q", cfg.ServerAddress)
	}
	if cfg.ReportInterval != 10 {
		t.Errorf("unexpected ReportInterval: %d", cfg.ReportInterval)
	}
	if cfg.PollInterval != 5 {
		t.Errorf("unexpected PollInterval: %d", cfg.PollInterval)
	}
	if cfg.RateLimit != 42 {
		t.Errorf("unexpected RateLimit: %d", cfg.RateLimit)
	}
	if cfg.HashKey != "hashkey" {
		t.Errorf("unexpected HashKey: %q", cfg.HashKey)
	}
	if cfg.PublicCryptoKeyPath != "/tmp/pub.key" {
		t.Errorf("unexpected PublicCryptoKeyPath: %q", cfg.PublicCryptoKeyPath)
	}
	if cfg.JSONConfigFile != "/tmp/config.json" {
		t.Errorf("unexpected JSONConfigFile: %q", cfg.JSONConfigFile)
	}
}

func TestParseEnv_ServerConfig(t *testing.T) {
	// временно сохраняем и восстанавливаем переменные окружения
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range origEnv {
			parts := splitEnv(kv)
			os.Setenv(parts[0], parts[1])
		}
	}()

	// устанавливаем тестовые переменные окружения
	os.Setenv("ADDRESS", "0.0.0.0:9090")
	os.Setenv("STORE_INTERVAL", "15")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/store")
	os.Setenv("RESTORE", "true")
	os.Setenv("DATABASE_DSN", "postgres://u:p@/db")
	os.Setenv("KEY", "serverkey")
	os.Setenv("AUDIT_FILE", "/tmp/audit.log")
	os.Setenv("AUDIT_URL", "https://audit.example")
	os.Setenv("CRYPTO_KEY", "/tmp/priv.key")
	os.Setenv("CONFIG", "/tmp/server.json")

	cfg := &ServerConfig{}
	if err := parseEnv(cfg); err != nil {
		t.Fatalf("parseEnv failed: %v", err)
	}

	if cfg.ServerAddress != "0.0.0.0:9090" {
		t.Errorf("unexpected ServerAddress: %q", cfg.ServerAddress)
	}
	if cfg.StoreInterval != 15 {
		t.Errorf("unexpected StoreInterval: %d", cfg.StoreInterval)
	}
	if cfg.FileStoragePath != "/tmp/store" {
		t.Errorf("unexpected FileStoragePath: %q", cfg.FileStoragePath)
	}
	if !cfg.RestoreMetricsFromFile {
		t.Errorf("unexpected RestoreMetricsFromFile: %v", cfg.RestoreMetricsFromFile)
	}
	if cfg.DatabaseDSN != "postgres://u:p@/db" {
		t.Errorf("unexpected DatabaseDSN: %q", cfg.DatabaseDSN)
	}
	if cfg.HashKey != "serverkey" {
		t.Errorf("unexpected HashKey: %q", cfg.HashKey)
	}
	if cfg.AuditFile != "/tmp/audit.log" {
		t.Errorf("unexpected AuditFile: %q", cfg.AuditFile)
	}
	if cfg.AuditURL != "https://audit.example" {
		t.Errorf("unexpected AuditURL: %q", cfg.AuditURL)
	}
	if cfg.PrivateCryptoKeyPath != "/tmp/priv.key" {
		t.Errorf("unexpected PrivateCryptoKeyPath: %q", cfg.PrivateCryptoKeyPath)
	}
	if cfg.JSONConfigFile != "/tmp/server.json" {
		t.Errorf("unexpected JSONConfigFile: %q", cfg.JSONConfigFile)
	}
}

// helper to split key=value from os.Environ()
func splitEnv(kv string) [2]string {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			return [2]string{kv[:i], kv[i+1:]}
		}
	}
	return [2]string{kv, ""}
}
