package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"github.com/MKhiriev/stunning-adventure/internal/misc"
	"log"
)

func main() {
	misc.ParseAgentFlags()
	err := agent.NewMetricsAgent(misc.ServerAddress.String(), "update", misc.ReportInterval, misc.PollInterval).Run()
	log.Fatal(err)
}
