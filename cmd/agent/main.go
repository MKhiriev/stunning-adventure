package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/misc"
	"github.com/caarlos0/env/v11"
	"log"
)

type AgentConfig struct {
	ServerAddress  *string `env:"ADDRESS"`
	ReportInterval *int64  `env:"REPORT_INTERVAL"`
	PollInterval   *int64  `env:"POLL_INTERVAL"`
}

func main() {
	cfg := Init()
	err := agent.NewMetricsAgent(*cfg.ServerAddress, "update", *cfg.ReportInterval, *cfg.PollInterval).Run()
	log.Fatal(err)
}

func Init() AgentConfig {
	var cfg AgentConfig
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	// if all values are not nil return cfg
	if cfg.ServerAddress != nil && cfg.ReportInterval != nil && cfg.PollInterval != nil {
		return cfg
	}

	misc.ParseAgentFlags()

	if cfg.ServerAddress == nil {
		commandLineServerAddress := misc.ServerAddress.String()
		cfg.ServerAddress = &commandLineServerAddress
	}
	if cfg.ReportInterval == nil {
		commandLineReportInterval := misc.ReportInterval
		cfg.ReportInterval = &commandLineReportInterval
	}
	if cfg.PollInterval == nil {
		commandLinePollInterval := misc.PollInterval
		cfg.PollInterval = &commandLinePollInterval
	}

	return cfg
}
