package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
)

func main() {
	cfg := config.GetAgentConfigs()
	log := logger.NewLogger("metrics-agent")
	log.Debug().Any("cfg-agent", cfg).Msg("")
	log.Info().Msg("Agent started")

	err := agent.NewMetricsAgent("updates", cfg, log).Run()
	log.Err(err).Caller().Str("func", "main").Msg("error occurred in agent during running")
}
