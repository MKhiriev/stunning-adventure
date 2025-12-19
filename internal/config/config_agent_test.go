package config

import (
	"testing"
)

func TestAgentConfigSetDefault(t *testing.T) {
	cfg := AgentConfig{}
	cfg.setDefault()

	if cfg.ServerAddress != defaultServerAddress {
		t.Errorf("setDefault() ServerAddress = %q, want %q", cfg.ServerAddress, defaultServerAddress)
	}
	if cfg.ReportInterval != defaultReportInterval {
		t.Errorf("setDefault() ReportInterval = %d, want %d", cfg.ReportInterval, defaultReportInterval)
	}
	if cfg.PollInterval != defaultPollInterval {
		t.Errorf("setDefault() PollInterval = %d, want %d", cfg.PollInterval, defaultPollInterval)
	}
	if cfg.RateLimit != defaultRateLimit {
		t.Errorf("setDefault() RateLimit = %d, want %d", cfg.RateLimit, defaultRateLimit)
	}

	// проверяем, что не перезаписывает существующие значения
	cfg2 := AgentConfig{
		ServerAddress:  "10.0.0.1:8080",
		ReportInterval: 5,
		PollInterval:   3,
		RateLimit:      100,
	}
	cfg2.setDefault()
	if cfg2.ServerAddress != "10.0.0.1:8080" {
		t.Errorf("setDefault() modified ServerAddress unexpectedly")
	}
	if cfg2.ReportInterval != 5 {
		t.Errorf("setDefault() modified ReportInterval unexpectedly")
	}
	if cfg2.PollInterval != 3 {
		t.Errorf("setDefault() modified PollInterval unexpectedly")
	}
	if cfg2.RateLimit != 100 {
		t.Errorf("setDefault() modified RateLimit unexpectedly")
	}
}

func TestAgentConfigIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		cfg  AgentConfig
		want bool
	}{
		{"empty config", AgentConfig{}, true},
		{"only ServerAddress", AgentConfig{ServerAddress: "127.0.0.1:8080"}, true},
		{"only ReportInterval", AgentConfig{ReportInterval: 10}, true},
		{"only PollInterval", AgentConfig{PollInterval: 5}, true},
		{"only RateLimit", AgentConfig{RateLimit: 50}, true},
		{"all set", AgentConfig{
			ServerAddress:  "127.0.0.1:8080",
			ReportInterval: 10,
			PollInterval:   5,
			RateLimit:      50,
		}, false},
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
