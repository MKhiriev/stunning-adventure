package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"log"
)

func main() {
	cfg := Init()
	err := agent.NewMetricsAgent("update", cfg).Run()
	log.Fatal(err)
}

func Init() *config.AgentConfig {
	return config.GetAgentConfigs()
}
