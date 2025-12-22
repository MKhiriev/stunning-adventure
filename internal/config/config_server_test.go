package config

import (
	"testing"
)

func TestServerConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServerConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     ServerConfig{ServerAddress: "127.0.0.1:8080", StoreInterval: 10},
			wantErr: false,
		},
		{
			name:    "missing ServerAddress",
			cfg:     ServerConfig{StoreInterval: 10},
			wantErr: true,
		},
		{
			name:    "missing StoreInterval",
			cfg:     ServerConfig{ServerAddress: "127.0.0.1:8080"},
			wantErr: true,
		},
		{
			name:    "both missing",
			cfg:     ServerConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerConfigSetDefault(t *testing.T) {
	cfg := ServerConfig{}
	cfg.setDefault()

	if cfg.ServerAddress != defaultServerAddress {
		t.Errorf("setDefault() ServerAddress = %q, want %q", cfg.ServerAddress, defaultServerAddress)
	}
	if cfg.StoreInterval != defaultStoreInterval {
		t.Errorf("setDefault() StoreInterval = %d, want %d", cfg.StoreInterval, defaultStoreInterval)
	}

	// проверяем, что не переписывает существующие значения
	cfg2 := ServerConfig{ServerAddress: "1.2.3.4:9000", StoreInterval: 5}
	cfg2.setDefault()
	if cfg2.ServerAddress != "1.2.3.4:9000" {
		t.Errorf("setDefault() modified ServerAddress unexpectedly")
	}
	if cfg2.StoreInterval != 5 {
		t.Errorf("setDefault() modified StoreInterval unexpectedly")
	}
}

func TestServerConfigIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		cfg  ServerConfig
		want bool
	}{
		{"empty config", ServerConfig{}, true},
		{"only ServerAddress", ServerConfig{ServerAddress: "127.0.0.1:8080"}, true},
		{"only StoreInterval", ServerConfig{StoreInterval: 10}, true},
		{"both set", ServerConfig{ServerAddress: "127.0.0.1:8080", StoreInterval: 10}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.isEmpty()
			if got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServerConfig_DefaultsApplied(t *testing.T) {
	resetFlags()
	clearEnv(t, "ADDRESS", "STORE_INTERVAL", "CONFIG")

	cfg, err := GetServerConfigs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != defaultServerAddress {
		t.Fatalf("expected default address, got %s", cfg.ServerAddress)
	}

	if cfg.StoreInterval != defaultStoreInterval {
		t.Fatalf("expected default store interval, got %d", cfg.StoreInterval)
	}
}

func TestServerConfig_JSONUsed_WhenFlagSpecified(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "json:8080",
		"store_interval": "10s"
	}`)

	resetFlags("-c", file)
	clearEnv(t, "CONFIG")

	cfg, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "json:8080" {
		t.Fatalf("expected json address, got %s", cfg.ServerAddress)
	}

	if cfg.StoreInterval != 10 {
		t.Fatalf("expected json store interval, got %d", cfg.StoreInterval)
	}
}

func TestServerConfig_FlagsOverrideJSON(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "127.0.0.1:8080",
		"store_interval": "10s"
	}`)

	resetFlags(
		"-c", file,
		"-a", "127.0.0.1:9090",
		"-i", "3",
	)
	clearEnv(t, "CONFIG")

	cfg, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:9090" {
		t.Fatalf("flag should override json, got %s", cfg.ServerAddress)
	}

	if cfg.StoreInterval != 3 {
		t.Fatalf("flag store_interval should override json")
	}
}

func TestServerConfig_EnvOverridesAll(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "127.0.0.1:8080",
		"store_interval": "10s"
	}`)

	resetFlags(
		"-c", file,
		"-a", "127.0.0.1:9090",
		"-i", "3",
	)

	t.Setenv("ADDRESS", "127.0.0.1:7070")
	t.Setenv("STORE_INTERVAL", "20")

	cfg, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:7070" {
		t.Fatalf("env should override all, got %s", cfg.ServerAddress)
	}

	if cfg.StoreInterval != 20 {
		t.Fatalf("env should override all store interval")
	}
}

func TestServerConfig_JSONFromEnvCONFIG(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "127.0.0.1:8080",
		"store_interval": "15s"
	}`)

	resetFlags()
	t.Setenv("CONFIG", file)

	cfg, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.1:8080" {
		t.Fatalf("expected json address from env")
	}
}

func TestServerConfig_InvalidJSON_ReturnsError(t *testing.T) {
	file := writeTempJSON(t, `{ invalid json`)

	resetFlags("-c", file)
	clearEnv(t, "CONFIG")

	_, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestServerConfig_ValidationFails(t *testing.T) {
	file := writeTempJSON(t, `{
		"store_interval": "10s"
	}`)

	resetFlags("-c", file)
	clearEnv(t, "CONFIG")

	cfg, err := newServerConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults().
		build()
	if err != nil {
		t.Fatalf("expected validation error")
	}
	if cfg.StoreInterval != 10 {
		t.Fatalf("expected store interval %v from JSON", 10)
	}
}
