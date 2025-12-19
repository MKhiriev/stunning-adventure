package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// Test GetAgentConfigs when -config points to JSON file: JSON values should be loaded
func TestGetAgentConfigs_JSONFallback(t *testing.T) {
	// create temp json file
	dir := t.TempDir()
	file := filepath.Join(dir, "agent.json")

	// JSON must match agentJSONConfig structure:
	// report_interval and poll_interval are strings because they map to utils.Duration
	payload := map[string]any{
		"address":         "1.2.3.4:8080",
		"report_interval": "5s",
		"poll_interval":   "2s",
		"rate_limit":      7,
		"hash_key":        "hk",
		"crypto_key":      "pub.pem",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(file, b, 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	runWithTempEnvAndArgs(t, []string{"-config", file}, nil, func() {
		cfg, err := GetAgentConfigs()
		if err != nil {
			t.Fatalf("GetAgentConfigs returned error: %v", err)
		}

		if cfg.ServerAddress != "1.2.3.4:8080" {
			t.Fatalf("ServerAddress mismatch: %q", cfg.ServerAddress)
		}
		if cfg.ReportInterval != int64(5) {
			t.Fatalf("ReportInterval mismatch: %d", cfg.ReportInterval)
		}
		if cfg.PollInterval != int64(2) {
			t.Fatalf("PollInterval mismatch: %d", cfg.PollInterval)
		}
		if cfg.RateLimit != 7 {
			t.Fatalf("RateLimit mismatch: %d", cfg.RateLimit)
		}
		if cfg.HashKey != "hk" {
			t.Fatalf("HashKey mismatch: %q", cfg.HashKey)
		}
		if cfg.PublicCryptoKeyPath != "pub.pem" {
			t.Fatalf("PublicCryptoKeyPath mismatch: %q", cfg.PublicCryptoKeyPath)
		}
	})
}

// Test GetServerConfigs when -config points to JSON file: JSON values should be loaded
func TestGetServerConfigs_JSONFallback(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "server.json")

	payload := map[string]any{
		"address":        "0.0.0.0:9090",
		"store_interval": "300s",
		"store_file":     "/tmp/store",
		"restore":        true,
		"database_dsn":   "postgres://u:p@/db",
		"hash_key":       "hk2",
		"audit_file":     "audit.log",
		"audit_url":      "http://audit",
		"crypto_key":     "priv.pem",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(file, b, 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	runWithTempEnvAndArgs(t, []string{"-config", file}, nil, func() {
		cfg, err := GetServerConfigs()
		if err != nil {
			t.Fatalf("GetServerConfigs returned error: %v", err)
		}

		// store_interval "300s" -> 300
		if cfg.StoreInterval != int64(300) {
			t.Fatalf("StoreInterval mismatch: %d", cfg.StoreInterval)
		}
		if cfg.ServerAddress != "0.0.0.0:9090" {
			t.Fatalf("ServerAddress mismatch: %q", cfg.ServerAddress)
		}
		if cfg.FileStoragePath != "/tmp/store" {
			t.Fatalf("FileStoragePath mismatch: %q", cfg.FileStoragePath)
		}
		if !cfg.RestoreMetricsFromFile {
			t.Fatalf("RestoreMetricsFromFile expected true")
		}
		if cfg.DatabaseDSN != "postgres://u:p@/db" {
			t.Fatalf("DatabaseDSN mismatch: %q", cfg.DatabaseDSN)
		}
		if cfg.HashKey != "hk2" {
			t.Fatalf("HashKey mismatch: %q", cfg.HashKey)
		}
		if cfg.AuditFile != "audit.log" {
			t.Fatalf("AuditFile mismatch: %q", cfg.AuditFile)
		}
		if cfg.AuditURL != "http://audit" {
			t.Fatalf("AuditURL mismatch: %q", cfg.AuditURL)
		}
		if cfg.PrivateCryptoKeyPath != "priv.pem" {
			t.Fatalf("PrivateCryptoKeyPath mismatch: %q", cfg.PrivateCryptoKeyPath)
		}
	})
}

// Test that flags values (non-JSON) are preferred over JSON (when both provided via flags)
// we provide flags for address and interval and ensure they arrive in cfg (no JSON)
func TestGetServerConfigs_FlagsPreferredOverJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "server.json")

	// JSON contains different address to ensure flags win before JSON parsing would even run
	payload := map[string]any{
		"address":        "10.10.10.10:1",
		"store_interval": "1s",
	}
	b, _ := json.Marshal(payload)
	_ = os.WriteFile(file, b, 0644)

	// Provide flags directly (address and interval) — parseAgentFlags will set those and fillEmptyServerConfigParams will copy them
	runWithTempEnvAndArgs(t, []string{"-a", "8.8.8.8:5555", "-i", "77", "-config", file}, nil, func() {
		cfg, err := GetServerConfigs()
		if err != nil {
			t.Fatalf("GetServerConfigs returned error: %v", err)
		}

		if cfg.ServerAddress != "8.8.8.8:5555" {
			t.Fatalf("ServerAddress mismatch: %q", cfg.ServerAddress)
		}
		if cfg.StoreInterval != 77 {
			t.Fatalf("StoreInterval mismatch: %d", cfg.StoreInterval)
		}
	})
}

// helper: temporarily replace flag.CommandLine and os.Args, clear env; restore after fn
func runWithTempEnvAndArgs(t *testing.T, args []string, env map[string]string, fn func()) {
	t.Helper()

	// save state
	origArgs := os.Args
	origEnv := os.Environ()
	origFlag := flag.CommandLine

	// restore
	defer func() {
		flag.CommandLine = origFlag
		os.Args = origArgs

		// restore env
		os.Clearenv()
		for _, kv := range origEnv {
			// split
			for i := 0; i < len(kv); i++ {
				if kv[i] == '=' {
					_ = os.Setenv(kv[:i], kv[i+1:])
					break
				}
			}
		}
	}()

	// clear env and set provided ones
	os.Clearenv()
	for k, v := range env {
		_ = os.Setenv(k, v)
	}

	// new FlagSet so flag.Parse() inside library reads our args
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = append([]string{"cmd"}, args...)

	fn()
}
