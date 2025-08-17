package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.GetServerConfigs()
	if err != nil {
		log.Err(err).Msg("invalid server configuration was passed")
		return
	}
	log := logger.NewLogger("metrics-server")
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

	handler := handlers.NewHandler(memStorage, fileStorage, conn, log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg)
}
