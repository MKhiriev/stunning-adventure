package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/store"
)

func main() {
	cfg := config.GetServerConfigs()
	log := logger.NewLogger("metrics-server")
	log.Info().Any("cfg-srv", cfg).Msg("Server started")

	conn, err := store.NewConnectPostgres(cfg, log)
	if err != nil {
		log.Err(err).Msg("connection to database failed")
		return
	}
	memStorage := store.NewMemStorage()
	fileStorage := store.NewFileStorage(memStorage, cfg)
	handler := handlers.NewHandler(memStorage, fileStorage, conn, log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg)
}
