package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMetricsService is a mock implementation of service.MetricsService
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Save(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(models.Metrics), args.Error(1)
}

func (m *MockMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func (m *MockMetricsService) Get(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	args := m.Called(ctx, metric)
	return args.Get(0).(models.Metrics), args.Error(1)
}

func (m *MockMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Metrics), args.Error(1)
}

func newTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

func newTestConfig() *config.ServerConfig {
	return &config.ServerConfig{}
}

// ========== NewMetricsServer Tests ==========

func TestNewMetricsServer_Success(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)

	require.NoError(t, err)
	require.NotNil(t, server)
	assert.NotNil(t, server.metricsService)
	assert.NotNil(t, server.converter)
	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.cfg)
}

func TestNewMetricsServer_WithNilDependencies(t *testing.T) {
	t.Run("nil service", func(t *testing.T) {
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(nil, cfg, logger)

		require.NoError(t, err, "constructor allows nil service")
		assert.Nil(t, server.metricsService)
	})

	t.Run("nil logger", func(t *testing.T) {
		mockService := new(MockMetricsService)
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, nil)

		require.NoError(t, err, "constructor allows nil logger")
		assert.Nil(t, server.logger)
	})

	t.Run("nil config", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()

		server, err := NewMetricsServer(mockService, nil, logger)

		require.NoError(t, err, "constructor allows nil config")
		assert.Nil(t, server.cfg)
	})
}

// ========== UpdateMetrics Tests ==========

func TestUpdateMetrics_Success_SingleCounter(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("requests")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(42)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	// Expect SaveAll to be called with converted metrics
	mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
		return len(metrics) == 1 &&
			metrics[0].ID == "requests" &&
			metrics[0].MType == models.Counter &&
			*metrics[0].Delta == 42
	})).Return(nil)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Success_SingleGauge(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("temperature")
	protoMetric.SetType(proto.Metric_GAUGE)
	protoMetric.SetValue(36.6)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
		return len(metrics) == 1 &&
			metrics[0].ID == "temperature" &&
			metrics[0].MType == models.Gauge &&
			*metrics[0].Value == 36.6
	})).Return(nil)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Success_MultipleMetrics(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric1 := &proto.Metric{}
	protoMetric1.SetId("counter1")
	protoMetric1.SetType(proto.Metric_COUNTER)
	protoMetric1.SetDelta(10)

	protoMetric2 := &proto.Metric{}
	protoMetric2.SetId("gauge1")
	protoMetric2.SetType(proto.Metric_GAUGE)
	protoMetric2.SetValue(3.14)

	protoMetric3 := &proto.Metric{}
	protoMetric3.SetId("counter2")
	protoMetric3.SetType(proto.Metric_COUNTER)
	protoMetric3.SetDelta(20)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric1, protoMetric2, protoMetric3})

	mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
		return len(metrics) == 3
	})).Return(nil)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.NoError(t, err)
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_NilRequest(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	response, err := server.UpdateMetrics(context.Background(), nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, errNilRequest)
	assert.Nil(t, response)
	mockService.AssertNotCalled(t, "SaveAll")
}

func TestUpdateMetrics_Error_NilMetrics(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	request := &proto.UpdateMetricsRequest{}
	// Don't set metrics, leaving it nil

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errNoMetricsProvided)
	assert.Nil(t, response)
	mockService.AssertNotCalled(t, "SaveAll")
}

func TestUpdateMetrics_Error_EmptyMetricsSlice(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{})

	mockService.On("SaveAll", mock.Anything, []models.Metrics{}).Return(nil)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.NoError(t, err, "empty slice should be allowed")
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_ValidationError_EmptyID(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("valid_metric")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(validators.ErrEmptyID)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidMetrics)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_ValidationError_EmptyType(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("metric1")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(validators.ErrEmptyType)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidMetrics)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_ValidationError_NoValue(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("metric1")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(validators.ErrNoValue)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidMetrics)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_ValidationError_InvalidType(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("metric1")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(validators.ErrInvalidType)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errInvalidMetrics)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_Error_UnexpectedServiceError(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("metric1")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	unexpectedErr := errors.New("database connection failed")
	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(unexpectedErr)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.Error(t, err)
	assert.ErrorIs(t, err, errUnexpectedError)
	assert.Nil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_ContextPropagation(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("metric1")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	type testKey string

	ctx := context.WithValue(context.Background(), testKey("test_key"), "test_value")

	mockService.On("SaveAll", mock.MatchedBy(func(c context.Context) bool {
		return c.Value(testKey("test_key")) == "test_value"
	}), mock.Anything).Return(nil)

	response, err := server.UpdateMetrics(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_ZeroValues(t *testing.T) {
	t.Run("counter with zero delta", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("zero_counter")
		protoMetric.SetType(proto.Metric_COUNTER)
		protoMetric.SetDelta(0)

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
			return len(metrics) == 1 && *metrics[0].Delta == 0
		})).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})

	t.Run("gauge with zero value", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("zero_gauge")
		protoMetric.SetType(proto.Metric_GAUGE)
		protoMetric.SetValue(0.0)

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
			return len(metrics) == 1 && *metrics[0].Value == 0.0
		})).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateMetrics_NegativeValues(t *testing.T) {
	t.Run("counter with negative delta", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("negative_counter")
		protoMetric.SetType(proto.Metric_COUNTER)
		protoMetric.SetDelta(-100)

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
			return len(metrics) == 1 && *metrics[0].Delta == -100
		})).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})

	t.Run("gauge with negative value", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("negative_gauge")
		protoMetric.SetType(proto.Metric_GAUGE)
		protoMetric.SetValue(-273.15)

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
			return len(metrics) == 1 && *metrics[0].Value == -273.15
		})).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateMetrics_EmptyID(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	protoMetric := &proto.Metric{}
	protoMetric.SetId("")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	request := &proto.UpdateMetricsRequest{}
	request.SetMetrics([]*proto.Metric{protoMetric})

	mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
		return len(metrics) == 1 && metrics[0].ID == ""
	})).Return(nil)

	response, err := server.UpdateMetrics(context.Background(), request)

	require.NoError(t, err, "empty ID should be allowed at gRPC level")
	require.NotNil(t, response)
	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_LargeValues(t *testing.T) {
	t.Run("large counter", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("large_counter")
		protoMetric.SetType(proto.Metric_COUNTER)
		protoMetric.SetDelta(9223372036854775807) // max int64

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.Anything).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})

	t.Run("large gauge", func(t *testing.T) {
		mockService := new(MockMetricsService)
		logger := newTestLogger()
		cfg := newTestConfig()

		server, err := NewMetricsServer(mockService, cfg, logger)
		require.NoError(t, err)

		protoMetric := &proto.Metric{}
		protoMetric.SetId("large_gauge")
		protoMetric.SetType(proto.Metric_GAUGE)
		protoMetric.SetValue(1.7976931348623157e+308) // near max float64

		request := &proto.UpdateMetricsRequest{}
		request.SetMetrics([]*proto.Metric{protoMetric})

		mockService.On("SaveAll", mock.Anything, mock.Anything).Return(nil)

		response, err := server.UpdateMetrics(context.Background(), request)

		require.NoError(t, err)
		require.NotNil(t, response)
		mockService.AssertExpectations(t)
	})
}

func TestUpdateMetrics_ConcurrentRequests(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	const numGoroutines = 50

	mockService.On("SaveAll", mock.Anything, mock.Anything).Return(nil)

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			protoMetric := &proto.Metric{}
			protoMetric.SetId("concurrent_metric")
			protoMetric.SetType(proto.Metric_COUNTER)
			protoMetric.SetDelta(int64(id))

			request := &proto.UpdateMetricsRequest{}
			request.SetMetrics([]*proto.Metric{protoMetric})

			response, err := server.UpdateMetrics(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, response)

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	mockService.AssertExpectations(t)
}

func TestUpdateMetrics_SpecialCharactersInID(t *testing.T) {
	mockService := new(MockMetricsService)
	logger := newTestLogger()
	cfg := newTestConfig()

	server, err := NewMetricsServer(mockService, cfg, logger)
	require.NoError(t, err)

	specialIDs := []string{
		"metric-with-dashes",
		"metric_with_underscores",
		"metric.with.dots",
		"metric:with:colons",
		"метрика_кириллица",
		"metric with spaces",
	}

	for _, id := range specialIDs {
		t.Run(id, func(t *testing.T) {
			protoMetric := &proto.Metric{}
			protoMetric.SetId(id)
			protoMetric.SetType(proto.Metric_COUNTER)
			protoMetric.SetDelta(10)

			request := &proto.UpdateMetricsRequest{}
			request.SetMetrics([]*proto.Metric{protoMetric})

			mockService.On("SaveAll", mock.Anything, mock.MatchedBy(func(metrics []models.Metrics) bool {
				return len(metrics) == 1 && metrics[0].ID == id
			})).Return(nil).Once()

			response, err := server.UpdateMetrics(context.Background(), request)

			require.NoError(t, err)
			require.NotNil(t, response)
		})
	}

	mockService.AssertExpectations(t)
}
