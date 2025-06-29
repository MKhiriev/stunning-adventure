package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	cfg, log := Init()
	log.Info().Msg("Server started")

	handler := handlers.NewHandler(log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg.ServerAddress)
}

func Init() (*config.ServerConfig, *zerolog.Logger) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("role", "metrics-server").
		Logger()

	return config.GetServerConfigs(), &logger
}
