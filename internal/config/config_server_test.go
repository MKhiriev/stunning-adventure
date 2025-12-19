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
