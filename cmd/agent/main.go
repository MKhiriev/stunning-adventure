package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	cfg, log := Init()
	err := agent.NewMetricsAgent(cfg.ServerAddress, "update", cfg.ReportInterval, cfg.PollInterval, log).Run()
	log.Err(err)
}

func Init() (*config.AgentConfig, *zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", "metrics-agent").
		Logger()

	return config.GetAgentConfigs(), &logger
}
