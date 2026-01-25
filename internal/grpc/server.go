package grpc

import (
	"context"
	"errors"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/rs/zerolog"
)

type MetricsServer struct {
	proto.UnimplementedMetricsServer

	metricsService service.MetricsService
	converter      MetricsConverter

	logger *zerolog.Logger
	cfg    *config.ServerConfig // todo replace with essential configs
}

func NewMetricsServer(metricsService service.MetricsService, cfg *config.ServerConfig, logger *zerolog.Logger) (*MetricsServer, error) {
	return &MetricsServer{
		UnimplementedMetricsServer: proto.UnimplementedMetricsServer{},
		metricsService:             metricsService,
		converter:                  newMetricConverter(),
		logger:                     logger,
		cfg:                        cfg,
	}, nil
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, metricsRequest *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if metricsRequest == nil {
		s.logger.Err(errNilRequest).
			Str("func", "*grpc.MetricsServer.UpdateMetrics").
			Msg("request has no metrics")
		return nil, errNilRequest
	}

	s.logger.Info().Str("func", "*grpc.MetricsServer.UpdateMetrics").Msg("UpdateMetrics was called")

	protoMetrics := metricsRequest.GetMetrics()
	if protoMetrics == nil {
		s.logger.Err(errNoMetricsProvided).
			Str("func", "*grpc.MetricsServer.UpdateMetrics").
			Msg("error getting metrics from proto request struct")
		return nil, errNoMetricsProvided
	}

	metrics, err := s.converter.ConvertMetricsFromProto(protoMetrics...)
	if err != nil {
		s.logger.Err(err).
			Str("func", "*grpc.MetricsServer.UpdateMetrics").
			Msg("error converting proto metrics to regular metrics")
		return nil, errInvalidMetrics
	}

	// update all values + validation
	if err := s.metricsService.SaveAll(ctx, metrics); err != nil {
		switch {
		case errors.Is(err, validators.ErrEmptyID) || errors.Is(err, validators.ErrEmptyType) || errors.Is(err, validators.ErrNoValue) || errors.Is(err, validators.ErrInvalidType):
			s.logger.Err(err).Caller().Str("func", "*grpc.MetricsServer.UpdateMetrics").Msg("passed metric is not valid")
			return nil, errInvalidMetrics
		default:
			s.logger.Err(err).Caller().Str("func", "*grpc.MetricsServer.UpdateMetrics").Msg("error occurred during metric update")
			return nil, errUnexpectedError
		}
	}

	s.logger.Info().Str("func", "*grpc.MetricsServer.UpdateMetrics").Msg("metrics are updated!")
	return &proto.UpdateMetricsResponse{}, nil
}
