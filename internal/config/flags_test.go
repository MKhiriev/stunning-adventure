package config

import (
	"flag"
	"os"
	"testing"
)

func TestParseServerFlags_AllFlags(t *testing.T) {
	withTempFlags(t, []string{
		"-a", "10.0.0.1:7070",
		"-i", "15",
		"-f", "/tmp/store",
		"-r",
		"-d", "postgres://u:p@/db",
		"-k", "mykey",
		"-audit-file", "/tmp/audit.log",
		"-audit-url", "https://audit",
		"-crypto-key", "/tmp/priv.key",
		"-config", "/etc/app/config.json",
	}, func() {
		cfg := ParseServerFlags()
		if cfg.ServerAddress != "10.0.0.1:7070" {
			t.Fatalf("unexpected ServerAddress: %q", cfg.ServerAddress)
		}
		if cfg.StoreInterval != 15 {
			t.Fatalf("unexpected StoreInterval: %d", cfg.StoreInterval)
		}
		if cfg.FileStoragePath != "/tmp/store" {
			t.Fatalf("unexpected FileStoragePath: %q", cfg.FileStoragePath)
		}
		if cfg.RestoreMetricsFromFile != true {
			t.Fatalf("unexpected RestoreMetricsFromFile: %v", cfg.RestoreMetricsFromFile)
		}
		if cfg.DatabaseDSN != "postgres://u:p@/db" {
			t.Fatalf("unexpected DatabaseDSN: %q", cfg.DatabaseDSN)
		}
		if cfg.HashKey != "mykey" {
			t.Fatalf("unexpected HashKey: %q", cfg.HashKey)
		}
		if cfg.AuditFile != "/tmp/audit.log" {
			t.Fatalf("unexpected AuditFile: %q", cfg.AuditFile)
		}
		if cfg.AuditURL != "https://audit" {
			t.Fatalf("unexpected AuditURL: %q", cfg.AuditURL)
		}
		if cfg.PrivateCryptoKeyPath != "/tmp/priv.key" {
			t.Fatalf("unexpected PrivateCryptoKeyPath: %q", cfg.PrivateCryptoKeyPath)
		}
		if cfg.JSONConfigFile != "/etc/app/config.json" {
			t.Fatalf("unexpected JSONConfigFile: %q", cfg.JSONConfigFile)
		}
	})
}

func TestParseServerFlags_ShortConfigFlag(t *testing.T) {
	withTempFlags(t, []string{
		"-c", "/tmp/conf.json",
	}, func() {
		cfg := ParseServerFlags()
		if cfg.JSONConfigFile != "/tmp/conf.json" {
			t.Fatalf("expected JSONConfigFile from -c, got %q", cfg.JSONConfigFile)
		}
	})
}

func TestParseAgentFlags_AllFlags(t *testing.T) {
	withTempFlags(t, []string{
		"-a", "192.168.0.5:6060",
		"-p", "7",
		"-r", "11",
		"-k", "agentkey",
		"-l", "42",
		"-crypto-key", "/tmp/pub.pem",
		"-config", "/tmp/agent.json",
	}, func() {
		cfg := parseAgentFlags()
		if cfg.ServerAddress != "192.168.0.5:6060" {
			t.Fatalf("unexpected ServerAddress: %q", cfg.ServerAddress)
		}
		if cfg.PollInterval != 7 {
			t.Fatalf("unexpected PollInterval: %d", cfg.PollInterval)
		}
		if cfg.ReportInterval != 11 {
			t.Fatalf("unexpected ReportInterval: %d", cfg.ReportInterval)
		}
		if cfg.HashKey != "agentkey" {
			t.Fatalf("unexpected HashKey: %q", cfg.HashKey)
		}
		if cfg.RateLimit != 42 {
			t.Fatalf("unexpected RateLimit: %d", cfg.RateLimit)
		}
		if cfg.PublicCryptoKeyPath != "/tmp/pub.pem" {
			t.Fatalf("unexpected PublicCryptoKeyPath: %q", cfg.PublicCryptoKeyPath)
		}
		if cfg.JSONConfigFile != "/tmp/agent.json" {
			t.Fatalf("unexpected JSONConfigFile: %q", cfg.JSONConfigFile)
		}
	})
}

func TestNetAddress_Set_ValidIP(t *testing.T) {
	var a NetAddress
	if err := a.Set("127.0.0.1:8080"); err != nil {
		t.Fatalf("unexpected error for valid ip: %v", err)
	}
	if a.Host != "127.0.0.1" || a.Port != 8080 {
		t.Fatalf("unexpected parsed values: %+v", a)
	}
	if got := a.String(); got != "127.0.0.1:8080" {
		t.Fatalf("unexpected String(): %q", got)
	}
}

func TestNetAddress_Set_Localhost(t *testing.T) {
	var a NetAddress
	if err := a.Set("localhost:9000"); err != nil {
		t.Fatalf("unexpected error for localhost: %v", err)
	}
	if a.Host != "localhost" || a.Port != 9000 {
		t.Fatalf("unexpected parsed values: %+v", a)
	}
}

func TestNetAddress_Set_InvalidFormat(t *testing.T) {
	var a NetAddress
	if err := a.Set("no-colon-here"); err == nil {
		t.Fatal("expected error for missing colon, got nil")
	}
	if err := a.Set("too:many:parts:1"); err == nil {
		t.Fatal("expected error for too many parts, got nil")
	}
}

func TestNetAddress_Set_NonNumericPort(t *testing.T) {
	var a NetAddress
	if err := a.Set("127.0.0.1:port"); err == nil {
		t.Fatal("expected error for non-numeric port, got nil")
	}
}

func TestNetAddress_Set_PortOutOfRange(t *testing.T) {
	var a NetAddress
	if err := a.Set("127.0.0.1:0"); err == nil {
		t.Fatal("expected error for port 0 (non-positive), got nil")
	}
	if err := a.Set("127.0.0.1:-1"); err == nil {
		t.Fatal("expected error for negative port, got nil")
	}
}

func TestNetAddress_Set_InvalidIP(t *testing.T) {
	var a NetAddress
	if err := a.Set("notanip:8080"); err == nil {
		t.Fatal("expected error for invalid IP, got nil")
	}
}

func TestNetAddress_String_Empty(t *testing.T) {
	var a NetAddress
	if a.String() != "" {
		t.Fatalf("expected empty string for zero NetAddress, got %q", a.String())
	}
}

// helper to replace flag.CommandLine and os.Args for a test, and restore afterwards
func withTempFlags(t *testing.T, args []string, fn func()) {
	t.Helper()
	// save global state
	orig := flag.CommandLine
	origArgs := os.Args

	// restore at the end
	defer func() {
		flag.CommandLine = orig
		os.Args = origArgs
	}()

	// new flag set
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	// set os.Args so flag.Parse() inside target functions will parse these
	os.Args = append([]string{"cmd"}, args...)

	fn()
}
