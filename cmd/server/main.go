package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

func main() {
	cfg := config.GetServerConfigs()
	log := utils.NewLogger("metrics-server")
	log.Info().Msg("Server started")

	memStorage := store.NewMemStorage()
	fileStorage := store.NewFileStorage(memStorage, cfg)
	handler := handlers.NewHandler(memStorage, fileStorage, log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg)
}
