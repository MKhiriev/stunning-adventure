// Package main provides entry points for both the metrics server and metrics agent.
//
// The server main initializes configuration, logging, adapters, storage layers
// (database, file, memory cache), and the metrics service pipeline. It then starts
// an HTTP server to handle incoming metric collection and audit events.
//
// The agent main initializes configuration and logging, creates a MetricsAgent,
// and starts its routine to collect, cache, and send system metrics to the server.
package main

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/adapters"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
	ctx := context.Background()
	log := logger.NewLogger("metrics-server")
	cfg, err := config.GetServerConfigs()
	if err != nil {
		log.Err(err).Msg("invalid server configuration was passed")
		return
	}
	log.Debug().Any("cfg-srv", cfg).Send()
	log.Info().Msg("Server started")

	allAdapters := adapters.NewAdapters(cfg.AuditURL, log)

	memStorage := store.NewMemStorage(log)
	conn, err := store.NewConnectPostgres(ctx, cfg, log)
	if err != nil {
		log.Err(err).Msg("connection to database failed")
	}
	fileStorage, err := store.NewFileStorage(ctx, memStorage, cfg, log)
	if err != nil {
		log.Err(err).Msg("file storage creation failed")
	}

	metricsValidationService := service.NewValidatingMetricsService()
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
	auditService := service.NewAuditService(cfg.AuditFile, allAdapters.AuditEventAdapter, log)

	var privateKey *rsa.PrivateKey
	if cfg.PrivateCryptoKeyPath != "" {
		privateKey, err = utils.GetPrivateKey(cfg.PrivateCryptoKeyPath)
		if err != nil {
			log.Err(err).Msg("error occurred getting private key")
			return
		}
	}

	handler := handlers.NewHandler(metricsService, pingService, auditService, privateKey, cfg, log)
	myServer := new(server.Server)
	myServer.ServerRun(handler.Init(), cfg)
}

func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
