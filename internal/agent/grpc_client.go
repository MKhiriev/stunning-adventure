package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/grpc"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type GRPCMetricsSender struct {
	client  *grpc.Client
	address string
	localIP string
	logger  *zerolog.Logger
}

func NewGRPCMetricsSender(address string, retryConfig grpc.RetryConfig, logger *zerolog.Logger) (*GRPCMetricsSender, error) {
	localIP, err := utils.GetLocalIP(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP address: %w", err)
	}

	client, err := grpc.NewGRPCClient(address, retryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	return &GRPCMetricsSender{
		client:  client,
		address: address,
		localIP: localIP.String(),
		logger:  logger,
	}, nil
}

func (s *GRPCMetricsSender) Send(metrics ...models.Metrics) error {
	ctx := context.Background()

	ctx = grpc.AddIPToContext(ctx, s.localIP)
	s.logger.Debug().
		Str("ip", s.localIP).
		Msg("added local IP to request metadata")

	if err := s.client.UpdateMetrics(ctx, metrics); err != nil {
		return fmt.Errorf("failed to send metrics via grpc: %w", err)
	}

	s.logger.Info().
		Int("count", len(metrics)).
		Msg("successfully sent metrics via grpc")

	return nil
}

func (s *GRPCMetricsSender) Close() error {
	return s.client.Close()
}

func DefaultRetryConfig() grpc.RetryConfig {
	return grpc.RetryConfig{
		MaxAttempts: 3,
		Intervals: map[int]time.Duration{
			1: 1 * time.Second,
			2: 3 * time.Second,
			3: 5 * time.Second,
		},
	}
}
