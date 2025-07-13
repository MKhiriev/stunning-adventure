package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

func main() {
	cfg := config.GetAgentConfigs()
	log := utils.NewLogger("metrics-agent")
	log.Info().Msg("Agent started")

	err := agent.NewMetricsAgent("update", cfg, log).Run()
	log.Err(err).Caller().Str("func", "main").Msg("error occurred in agent during running")
}
