// Package main provides the entry point for the metrics agent application.
//
// The agent collects system metrics on the host machine, caches them in memory,
// and periodically sends them to the metrics server. It leverages the MetricsAgent
// abstraction for polling, batching, and sending metrics.
package main

import (
	"crypto/rsa"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	log := logger.NewLogger("metrics-agent")
	cfg, err := config.GetAgentConfigs()
	if err != nil {
		log.Err(err).Msg("error getting configs")
		return
	}

	log.Debug().Any("cfg-agent", cfg).Send()
	log.Info().Msg("Agent started")

	var publicKey *rsa.PublicKey
	if cfg.PublicCryptoKeyPath != "" {
		var err error
		publicKey, err = utils.GetPublicKey(cfg.PublicCryptoKeyPath)
		if err != nil {
			log.Err(err).Msg("error getting public key")
			return
		}
	}

	err = agent.NewMetricsAgent("updates", publicKey, cfg, log).Run()
	log.Err(err).Caller().Str("func", "main").Msg("error occurred in agent during running")
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
