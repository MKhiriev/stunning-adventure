package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/agent"
	"log"
)

func main() {
	err := agent.NewMetricsAgent("localhost:8080", "update").Run()
	log.Fatal(err)
}
