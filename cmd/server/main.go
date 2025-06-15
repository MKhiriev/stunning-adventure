package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"log"
)

func main() {
	cfg := Init()
	handler := handlers.NewHandler()
	myServer := new(server.Server)
	err := myServer.ServerRun(handler.Init(), cfg.ServerAddress)
	log.Fatal(err)
}

func Init() *config.ServerConfig {
	return config.GetServerConfigs()
}
