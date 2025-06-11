package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/misc"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/caarlos0/env/v11"
	"log"
)

type ServerConfig struct {
	ServerAddress *string `env:"ADDRESS"` // can define if anything was passed
}

func main() {
	cfg := Init()
	handler := handlers.NewHandler()
	myServer := new(server.Server)
	err := myServer.ServerRun(handler.Init(), *cfg.ServerAddress)
	log.Fatal(err)
}

func Init() ServerConfig {
	var cfg ServerConfig
	err := env.Parse(&cfg)

	if err != nil {
		log.Fatal(err)
	}

	// if ServerAddress is not nil return cfg
	if cfg.ServerAddress != nil {
		return cfg
	}

	// else get command line args
	misc.ParseServerFlags()

	// if command line args are empty server address will be assigned a default value
	if misc.ServerAddress.String() != "" {
		commandLineServerAddress := misc.ServerAddress.String()
		cfg.ServerAddress = &commandLineServerAddress
		return cfg
	}

	return cfg
}
