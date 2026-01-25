package agent

import (
	"fmt"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

// TODO implement
type GRPCMetricsSender struct {
	route string // server route for sending metrics
	// todo add gRPC client
	logger *zerolog.Logger

	realIP string
}

func NewGRPCMetricsSender(serverAddress, route string, retryIntervals map[int]time.Duration, cfg *config.AgentConfig, logger *zerolog.Logger) (*GRPCMetricsSender, error) {

	// dial server address
	realIp, err := utils.GetLocalIP(serverAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to get real agent IP: %w", err)
	}

	return &GRPCMetricsSender{
		route:  route,
		logger: logger,
		realIP: realIp.String(),
	}, nil
}

func (s *GRPCMetricsSender) Send(metrics ...models.Metrics) error {
	//TODO implement me
	panic("implement me")
}

func (s *GRPCMetricsSender) Close() error {
	//TODO implement me
	panic("implement me")
}
