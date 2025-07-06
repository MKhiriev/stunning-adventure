package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/services"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog"
	"os"
)

func main() {
	cfg, log := Init()
	log.Info().Msg("Server started")

	service := services.NewMetricsService(
		store.NewMemStorage(),
		store.NewFileStorage("metrics.log", cfg.FileStoragePath),
		cfg.StoreInterval,
		cfg.Restore,
	)
	handler := handlers.NewHandler(log, service)
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
