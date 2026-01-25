package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn      *grpc.ClientConn
	client    proto.MetricsClient
	converter MetricsConverter
	retry     RetryConfig
	logger    *zerolog.Logger
}

type RetryConfig struct {
	MaxAttempts int
	Intervals   map[int]time.Duration
}

func NewGRPCClient(address string, retry RetryConfig, logger *zerolog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Client{
		conn:      conn,
		client:    proto.NewMetricsClient(conn),
		converter: newMetricConverter(),
		retry:     retry,
		logger:    logger,
	}, nil
}

func (c *Client) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	protoMetrics, err := c.converter.ConvertMetricsToProto(metrics...)
	if err != nil {
		return fmt.Errorf("failed to convert metrics to proto: %w", err)
	}

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics(protoMetrics)

	return c.withRetry(ctx, func(ctx context.Context) error {
		_, updateMetricsErr := c.client.UpdateMetrics(ctx, request)
		return updateMetricsErr
	})
}

func (c *Client) withRetry(ctx context.Context, operation func(context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= c.retry.MaxAttempts; attempt++ {
		if attempt > 1 {
			interval := c.getRetryInterval(attempt)
			c.logger.Debug().
				Int("attempt", attempt).
				Dur("wait", interval).
				Msg("retrying grpc request")

			select {
			case <-time.After(interval):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		lastErr = operation(ctx)
		if lastErr == nil {
			return nil
		}

		c.logger.Err(lastErr).
			Int("attempt", attempt).
			Int("max_attempts", c.retry.MaxAttempts).
			Msg("grpc request failed")
	}

	return fmt.Errorf("grpc request failed after %d attempts: %w", c.retry.MaxAttempts, lastErr)
}

func (c *Client) getRetryInterval(attempt int) time.Duration {
	if interval, ok := c.retry.Intervals[attempt]; ok {
		return interval
	}
	return time.Second
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func AddIPToContext(ctx context.Context, ip string) context.Context {
	md := metadata.New(map[string]string{
		"x-real-ip": ip,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
