package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	cfg, log := Init()
	log.Info().Msg("Agent started")

	err := agent.NewMetricsAgent("update", cfg, log).Run()
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
