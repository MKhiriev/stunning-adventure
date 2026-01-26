package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ===== helpers =====

func newNopLogger() *zerolog.Logger {
	l := zerolog.Nop()
	return &l
}

func freeTCPAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())
	return addr
}

// ===== fake Metrics server =====

type fakeMetricsServer struct {
	proto.UnimplementedMetricsServer

	updateFn func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error)
}

func (s *fakeMetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, req)
	}
	return &proto.UpdateMetricsResponse{}, nil
}

// start test gRPC server and return address + cleanup
func startMetricsServer(t *testing.T, srv proto.MetricsServer) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	gs := gogrpc.NewServer()
	proto.RegisterMetricsServer(gs, srv)

	go func() {
		_ = gs.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		gs.Stop()
		_ = lis.Close()
	}
}

func newClientForTest(t *testing.T, address string, retry RetryConfig) *Client {
	t.Helper()
	conn, err := gogrpc.NewClient(address, gogrpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	return &Client{
		conn:      conn,
		client:    proto.NewMetricsClient(conn),
		converter: newMetricConverter(),
		retry:     retry,
		logger:    newNopLogger(),
	}
}

// ========== NewGRPCClient / Close Tests ==========

func TestNewGRPCClient_Success_AndClose(t *testing.T) {
	addr := freeTCPAddr(t) // we don't need a server for dialing with insecure; creating conn should succeed

	c, err := NewGRPCClient(addr, RetryConfig{MaxAttempts: 1, Intervals: map[int]time.Duration{}}, newNopLogger())
	require.NoError(t, err)
	require.NotNil(t, c)

	require.NotNil(t, c.conn)
	require.NotNil(t, c.client)
	require.NotNil(t, c.converter)

	err = c.Close()
	require.NoError(t, err)
}

func TestClient_Close_NilConn_NoError(t *testing.T) {
	c := &Client{conn: nil}
	require.NoError(t, c.Close())
}

// ========== AddIPToContext Tests ==========

func TestAddIPToContext_SetsOutgoingMetadata(t *testing.T) {
	ctx := AddIPToContext(context.Background(), "192.168.1.10")

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)

	values := md.Get("X-Real-IP")
	require.Len(t, values, 1)
	assert.Equal(t, "192.168.1.10", values[0])
}

// ========== getRetryInterval Tests ==========

func TestClient_getRetryInterval_CustomAndDefault(t *testing.T) {
	c := &Client{
		retry: RetryConfig{
			MaxAttempts: 3,
			Intervals: map[int]time.Duration{
				2: 25 * time.Millisecond,
			},
		},
	}

	assert.Equal(t, time.Second, c.getRetryInterval(1))
	assert.Equal(t, 25*time.Millisecond, c.getRetryInterval(2))
	assert.Equal(t, time.Second, c.getRetryInterval(3))
}

// ========== withRetry Tests ==========

func TestClient_withRetry_SuccessFirstAttempt(t *testing.T) {
	c := &Client{
		retry:  RetryConfig{MaxAttempts: 3, Intervals: map[int]time.Duration{2: 1 * time.Millisecond}},
		logger: newNopLogger(),
	}

	calls := 0
	err := c.withRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestClient_withRetry_SucceedsAfterRetry(t *testing.T) {
	c := &Client{
		retry: RetryConfig{
			MaxAttempts: 3,
			Intervals: map[int]time.Duration{
				2: 1 * time.Millisecond,
				3: 1 * time.Millisecond,
			},
		},
		logger: newNopLogger(),
	}

	calls := 0
	someErr := errors.New("temporary")

	err := c.withRetry(context.Background(), func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return someErr
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestClient_withRetry_ExhaustsAttempts_ReturnsWrappedError(t *testing.T) {
	c := &Client{
		retry:  RetryConfig{MaxAttempts: 3, Intervals: map[int]time.Duration{2: 1 * time.Millisecond, 3: 1 * time.Millisecond}},
		logger: newNopLogger(),
	}

	someErr := errors.New("always failing")
	calls := 0

	err := c.withRetry(context.Background(), func(ctx context.Context) error {
		calls++
		return someErr
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "grpc request failed after 3 attempts")
	assert.ErrorIs(t, err, someErr)
	assert.Equal(t, 3, calls)
}

func TestClient_withRetry_ContextCancelledBeforeSleep(t *testing.T) {
	c := &Client{
		retry: RetryConfig{
			MaxAttempts: 3,
			Intervals: map[int]time.Duration{
				2: 500 * time.Millisecond, // big; but context will cancel immediately
			},
		},
		logger: newNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := c.withRetry(ctx, func(ctx context.Context) error {
		calls++
		return errors.New("fail")
	})

	// attempt=1 will run immediately, attempt=2 will try to sleep and should return ctx.Err()
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls)
}

// ========== UpdateMetrics (integration-style with test server) Tests ==========

func TestClient_UpdateMetrics_Success_SendsRequest(t *testing.T) {
	var gotReq *proto.UpdateMetricsRequest

	srv := &fakeMetricsServer{
		updateFn: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			gotReq = req
			return &proto.UpdateMetricsResponse{}, nil
		},
	}

	addr, stop := startMetricsServer(t, srv)
	defer stop()

	c := newClientForTest(t, addr, RetryConfig{MaxAttempts: 1, Intervals: map[int]time.Duration{}})

	delta := int64(10)
	metrics := []models.Metrics{
		{ID: "requests", MType: models.Counter, Delta: &delta},
	}

	err := c.UpdateMetrics(context.Background(), metrics)
	require.NoError(t, err)

	require.NotNil(t, gotReq)
	require.NotNil(t, gotReq.GetMetrics())
	require.Len(t, gotReq.GetMetrics(), 1)
	assert.Equal(t, "requests", gotReq.GetMetrics()[0].GetId())
}

func TestClient_UpdateMetrics_AttachesIPMetadata(t *testing.T) {
	srv := &fakeMetricsServer{
		updateFn: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			require.True(t, ok)

			ips := md.Get("X-Real-IP")
			require.Len(t, ips, 1)
			assert.Equal(t, "192.168.1.77", ips[0])

			return &proto.UpdateMetricsResponse{}, nil
		},
	}

	addr, stop := startMetricsServer(t, srv)
	defer stop()

	c := newClientForTest(t, addr, RetryConfig{MaxAttempts: 1, Intervals: map[int]time.Duration{}})

	val := float64(1.5)
	metrics := []models.Metrics{
		{ID: "g", MType: models.Gauge, Value: &val},
	}

	ctx := AddIPToContext(context.Background(), "192.168.1.77")
	err := c.UpdateMetrics(ctx, metrics)
	require.NoError(t, err)
}

func TestClient_UpdateMetrics_RetryOnServerError(t *testing.T) {
	calls := 0

	srv := &fakeMetricsServer{
		updateFn: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			calls++
			if calls < 3 {
				return nil, errors.New("temporary server error")
			}
			return &proto.UpdateMetricsResponse{}, nil
		},
	}

	addr, stop := startMetricsServer(t, srv)
	defer stop()

	c := newClientForTest(t, addr, RetryConfig{
		MaxAttempts: 3,
		Intervals: map[int]time.Duration{
			2: 1 * time.Millisecond,
			3: 1 * time.Millisecond,
		},
	})

	delta := int64(1)
	metrics := []models.Metrics{
		{ID: "c", MType: models.Counter, Delta: &delta},
	}

	err := c.UpdateMetrics(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestClient_UpdateMetrics_ExhaustsRetries(t *testing.T) {
	srv := &fakeMetricsServer{
		updateFn: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			return nil, errors.New("always failing")
		},
	}

	addr, stop := startMetricsServer(t, srv)
	defer stop()

	c := newClientForTest(t, addr, RetryConfig{
		MaxAttempts: 3,
		Intervals: map[int]time.Duration{
			2: 1 * time.Millisecond,
			3: 1 * time.Millisecond,
		},
	})

	delta := int64(1)
	metrics := []models.Metrics{
		{ID: "c", MType: models.Counter, Delta: &delta},
	}

	err := c.UpdateMetrics(context.Background(), metrics)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "grpc request failed after 3 attempts")
}

func TestClient_UpdateMetrics_ContextDeadlineExceededDuringRetrySleep(t *testing.T) {
	calls := 0

	srv := &fakeMetricsServer{
		updateFn: func(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
			calls++
			return nil, errors.New("fail")
		},
	}

	addr, stop := startMetricsServer(t, srv)
	defer stop()

	c := newClientForTest(t, addr, RetryConfig{
		MaxAttempts: 3,
		Intervals: map[int]time.Duration{
			2: 250 * time.Millisecond, // longer than context deadline below
		},
	})

	val := float64(1.0)
	metrics := []models.Metrics{
		{ID: "g", MType: models.Gauge, Value: &val},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := c.UpdateMetrics(ctx, metrics)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// attempt=1 executed, then context deadline triggers during sleep for attempt=2
	assert.Equal(t, 1, calls)
}

func TestClient_UpdateMetrics_ErrorOnConversion(t *testing.T) {
	// If converter rejects metric (e.g., unsupported type), UpdateMetrics must return conversion error without dialing.
	c := &Client{
		conn:      nil,
		client:    nil,
		converter: newMetricConverter(),
		retry:     RetryConfig{MaxAttempts: 1, Intervals: map[int]time.Duration{}},
		logger:    newNopLogger(),
	}

	metrics := []models.Metrics{
		{ID: "x", MType: "unsupported"},
	}

	err := c.UpdateMetrics(context.Background(), metrics)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to convert metrics to proto")
}
