package config

import (
	"flag"
	"os"
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

func TestAgentConfig_JSON_NotUsed_WhenNotSpecified(t *testing.T) {
	resetFlags()
	clearEnv(t, "CONFIG")

	builder := newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults()

	cfg, err := builder.build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != defaultServerAddress {
		t.Fatalf("expected default address, got %s", cfg.ServerAddress)
	}
}

func TestAgentConfig_JSON_Used_WhenFlagSpecified(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "127.0.25.123:8080",
		"poll_interval": "10s",
		"report_interval": "20s",
		"rate_limit": 5
	}`)

	resetFlags("-c", file)
	clearEnv(t, "CONFIG")

	builder := newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults()

	cfg, err := builder.build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.25.123:8080" {
		t.Fatalf("expected json address, got %s", cfg.ServerAddress)
	}
}

func TestAgentConfig_FlagsOverrideJSON(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "json:8080",
		"poll_interval": "10s"
	}`)

	resetFlags(
		"-c", file,
		"-a", "127.0.25.123:9090",
		"-p", "3",
	)
	clearEnv(t, "CONFIG")

	builder := newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults()

	cfg, err := builder.build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.25.123:9090" {
		t.Fatalf("flag should override json")
	}

	if cfg.PollInterval != 3 {
		t.Fatalf("flag poll_interval should override json")
	}
}

func TestAgentConfig_EnvOverridesEverything(t *testing.T) {
	file := writeTempJSON(t, `{
		"address": "127.0.0.3:9090"
	}`)

	resetFlags("-c", file, "-a", "127.0.0.3:9090")

	t.Setenv("ADDRESS", "127.0.0.2:7070")

	builder := newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults()

	cfg, err := builder.build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddress != "127.0.0.2:7070" {
		t.Fatalf("env should override all, got %s", cfg.ServerAddress)
	}
}

func TestAgentConfig_DefaultsApplied(t *testing.T) {
	resetFlags()
	clearEnv(t, "ADDRESS", "CONFIG")

	builder := newAgentConfigBuilder().
		withEnv().
		withFlags().
		withJSON().
		withDefaults()

	cfg, err := builder.build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.PollInterval != defaultPollInterval {
		t.Fatalf("default poll interval not applied")
	}
}

func resetFlags(args ...string) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = append([]string{os.Args[0]}, args...)
}

func clearEnv(t *testing.T, keys ...string) {
	for _, k := range keys {
		t.Helper()
		_ = os.Unsetenv(k)
	}
}

func writeTempJSON(t *testing.T, content string) string {
	t.Helper()

	f, err := os.CreateTemp("", "cfg-*.json")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}

	_ = f.Close()
	return f.Name()
}
