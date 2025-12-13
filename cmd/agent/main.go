// Package main provides the entry point for the metrics agent application.
//
// The agent collects system metrics on the host machine, caches them in memory,
// and periodically sends them to the metrics server. It leverages the MetricsAgent
// abstraction for polling, batching, and sending metrics.
package main

import (
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	cfg := config.GetAgentConfigs()
	log := logger.NewLogger("metrics-agent")
	log.Debug().Any("cfg-agent", cfg).Send()
	log.Info().Msg("Agent started")

	err := agent.NewMetricsAgent("updates", cfg, log).Run()
	log.Err(err).Caller().Str("func", "main").Msg("error occurred in agent during running")
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}

	date := buildDate
	if date == "" {
		date = "N/A"
	}

	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}
