package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/store"
)

func main() {
	log := logger.NewLogger("metrics-server")
	cfg, err := config.GetServerConfigs()
	if err != nil {
		log.Err(err).Msg("invalid server configuration was passed")
		return
	}
	log.Info().Any("cfg-srv", cfg).Msg("Server started")

	memStorage := store.NewMemStorage(log)
	conn, err := store.NewConnectPostgres(cfg, log)
	if err != nil {
		log.Err(err).Msg("connection to database failed")
	}
	fileStorage, err := store.NewFileStorage(memStorage, cfg, log)
	if err != nil {
		log.Err(err).Msg("file storage creation failed")
	}

	metricsValidationService := service.NewValidatingMetricsService(log)
	metricsService, err := service.NewMetricsServiceBuilder(cfg, log).
		WithDB(conn).
		WithFile(fileStorage).
		WithCache(memStorage).
		WithWrapper(metricsValidationService).
		Build()
	if err != nil {
		log.Err(err).Msg("creation of metrics service failed")
		return
	}
	pingService, err := service.NewPingDBService(conn, log)
	if err != nil {
		log.Err(err).Msg("creation of ping db service failed")
		return
	}

	handler := handlers.NewHandler(metricsService, pingService, cfg, log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg)
}
