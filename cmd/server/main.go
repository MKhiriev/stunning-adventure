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
	"errors"
	"fmt"
	"net"

	"github.com/MKhiriev/stunning-adventure/internal/adapters"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	myGrpc "github.com/MKhiriev/stunning-adventure/internal/grpc"
	"github.com/MKhiriev/stunning-adventure/internal/handlers"
	"github.com/MKhiriev/stunning-adventure/internal/logger"
	myProto "github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/internal/server"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
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

	myServer := new(server.Server)
	setupServer(myServer, metricsService, pingService, auditService, privateKey, cfg, log)

	myServer.ServerRun()
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

func setupServer(server *server.Server, metricsService service.MetricsService, pingService service.PingService, auditService service.AuditPublisher, privateKey *rsa.PrivateKey, cfg *config.ServerConfig, log *zerolog.Logger) {
	if cfg.GrpcServerAddress == "" && cfg.ServerAddress == "" {
		log.Fatal().Msgf("no server was specified!")
	}

	if cfg.GrpcServerAddress != "" && cfg.GrpcServerAddress == cfg.ServerAddress {
		log.Fatal().Msgf("gRPC and HTTP servers have the same port")
	}

	if cfg.GrpcServerAddress != "" {
		grpcServer, lis, err := getGRPCServer(metricsService, cfg, log)
		if err != nil {
			log.Fatal().Msgf("grpc server setup fail: %v", err)
		}

		server.GRPCServer(grpcServer, lis)
	}
	if cfg.ServerAddress != "" {
		httpServerHandler, err := getHTTPServerHandler(metricsService, pingService, auditService, privateKey, cfg, log)
		if err != nil {
			log.Fatal().Msgf("http server setup fail: %v", err)
		}

		server.HTTPServer(httpServerHandler.Init(), cfg)
	}
}

func getHTTPServerHandler(metricsService service.MetricsService, pingService service.PingService, auditService service.AuditPublisher, privateKey *rsa.PrivateKey, cfg *config.ServerConfig, log *zerolog.Logger) (*handlers.Handler, error) {
	handler, err := handlers.NewHandler(metricsService, pingService, auditService, privateKey, cfg, log)
	if err != nil {
		return nil, err
	}

	return handler, nil

}

func getGRPCServer(metricsService service.MetricsService, cfg *config.ServerConfig, log *zerolog.Logger) (*grpc.Server, net.Listener, error) {
	if cfg.TrustedSubnet == "" {
		return nil, nil, errors.New("trusted subnet is empty")
	}

	grpcMetricsServer, err := myGrpc.NewMetricsServer(metricsService, cfg, log)
	if err != nil {
		return nil, nil, errors.New("error creating grpc server")
	}

	lis, err := net.Listen("tcp", cfg.GrpcServerAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	trustedSubnet, err := myGrpc.ParseTrustedSubnet(cfg.TrustedSubnet)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}

	var serverOpts []grpc.ServerOption
	if trustedSubnet != nil {
		log.Info().Str("trusted_subnet", trustedSubnet.String()).Msg("enabling trusted subnet check")
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(myGrpc.TrustedSubnetInterceptor(trustedSubnet, log)))
	}

	grpcServer := grpc.NewServer(serverOpts...)
	myProto.RegisterMetricsServer(grpcServer, grpcMetricsServer)

	return grpcServer, lis, nil
}
