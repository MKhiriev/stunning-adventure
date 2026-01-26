package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

func newTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

// ========== NewAuditAdapter Tests ==========

func TestNewAuditAdapter_Success(t *testing.T) {
	logger := newTestLogger()
	adapter := NewAuditAdapter("http://example.com", logger)

	require.NotNil(t, adapter)
	assert.IsType(t, &auditAdapter{}, adapter)
}

func TestNewAuditAdapter_EmptyRemoteServer_ReturnsNil(t *testing.T) {
	logger := newTestLogger()
	adapter := NewAuditAdapter("", logger)

	assert.Nil(t, adapter, "empty remote server should return nil")
}

func TestNewAuditAdapter_WithDifferentURLs(t *testing.T) {
	logger := newTestLogger()

	urls := []string{
		"http://localhost:8080",
		"https://api.example.com/audit",
		"http://192.168.1.1:9090/events",
		"https://audit.service.local/v1/events",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			adapter := NewAuditAdapter(url, logger)
			require.NotNil(t, adapter)

			aa, ok := adapter.(*auditAdapter)
			require.True(t, ok)
			assert.Equal(t, url, aa.remoteServer)
		})
	}
}

// ========== SendEvent Tests ==========

func TestAuditAdapter_SendEvent_Success(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)
	require.NotNil(t, adapter)

	event, err := models.NewAuditEvent("192.168.1.100", time.Unix(1705314645, 0), "metric1", "metric2")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, int64(1705314645), receivedEvent.TimeStamp)
	assert.Equal(t, []string{"metric1", "metric2"}, receivedEvent.Metrics)
	assert.Equal(t, "192.168.1.100", receivedEvent.IPAddress)
}

func TestAuditAdapter_SendEvent_SingleMetric(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("10.0.0.1", time.Unix(1000000000, 0), "single_metric")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, receivedEvent.Metrics, 1)
	assert.Equal(t, "single_metric", receivedEvent.Metrics[0])
}

func TestAuditAdapter_SendEvent_MultipleMetrics(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("172.16.0.1", time.Now(), "m1", "m2", "m3", "m4", "m5")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, receivedEvent.Metrics, 5)
	assert.Equal(t, []string{"m1", "m2", "m3", "m4", "m5"}, receivedEvent.Metrics)
}

func TestAuditAdapter_SendEvent_IPv6Address(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("2001:db8::1", time.Now(), "metric")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "2001:db8::1", receivedEvent.IPAddress)
}

func TestAuditAdapter_SendEvent_SpecialCharactersInMetricNames(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	specialMetrics := []string{
		"metric-with-dashes",
		"metric_with_underscores",
		"metric.with.dots",
		"метрика_кириллица",
		"metric with spaces",
	}

	event, err := models.NewAuditEvent("127.0.0.1", time.Now(), specialMetrics...)
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, specialMetrics, receivedEvent.Metrics)
}

func TestAuditAdapter_SendEvent_ServerReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
	require.NoError(t, err)

	// resty doesn't return error on non-2xx status codes by default
	// so this should succeed unless we configure error handling
	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err, "resty by default doesn't treat non-2xx as error")
}

func TestAuditAdapter_SendEvent_NetworkError(t *testing.T) {
	logger := newTestLogger()
	// Use invalid URL to trigger network error
	adapter := NewAuditAdapter("http://invalid-host-that-does-not-exist-12345.local", logger)

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotSent)
}

func TestAuditAdapter_SendEvent_ServerUnavailable(t *testing.T) {
	// Start server and immediately close it
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(serverURL, logger)

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotSent)
}

func TestAuditAdapter_SendEvent_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = adapter.SendEvent(ctx, event)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotSent)
}

func TestAuditAdapter_SendEvent_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = adapter.SendEvent(ctx, event)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEventNotSent)
}

func TestAuditAdapter_SendEvent_ConcurrentRequests(t *testing.T) {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			event, err := models.NewAuditEvent("192.168.1.1", time.Now(), "metric")
			require.NoError(t, err)

			err = adapter.SendEvent(context.Background(), event)
			require.NoError(t, err)

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	assert.Equal(t, numGoroutines, requestCount)
}

func TestAuditAdapter_SendEvent_LargeNumberOfMetrics(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	// Create 1000 metrics
	metrics := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		metrics[i] = "metric_" + string(rune(i))
	}

	event, err := models.NewAuditEvent("192.168.1.1", time.Now(), metrics...)
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Len(t, receivedEvent.Metrics, 1000)
}

func TestAuditAdapter_SendEvent_EmptyIPAddress(t *testing.T) {
	var receivedEvent models.AuditEvent

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := newTestLogger()
	adapter := NewAuditAdapter(server.URL, logger)

	event, err := models.NewAuditEvent("", time.Now(), "metric")
	require.NoError(t, err)

	err = adapter.SendEvent(context.Background(), event)
	require.NoError(t, err)

	assert.Equal(t, "", receivedEvent.IPAddress)
}

func TestAuditAdapter_SendEvent_DifferentTimestamps(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp time.Time
	}{
		{"epoch", time.Unix(0, 0)},
		{"year 2000", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"current time", time.Now()},
		{"future", time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var receivedEvent models.AuditEvent

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err := json.NewDecoder(r.Body).Decode(&receivedEvent)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			logger := newTestLogger()
			adapter := NewAuditAdapter(server.URL, logger)

			event, err := models.NewAuditEvent("192.168.1.1", tc.timestamp, "metric")
			require.NoError(t, err)

			err = adapter.SendEvent(context.Background(), event)
			require.NoError(t, err)

			assert.Equal(t, tc.timestamp.Unix(), receivedEvent.TimeStamp)
		})
	}
}

// ========== NewAdapters Tests ==========

func TestNewAdapters_Success(t *testing.T) {
	logger := newTestLogger()
	adapters := NewAdapters("http://example.com", logger)

	require.NotNil(t, adapters)
	assert.NotNil(t, adapters.AuditEventAdapter)
}

func TestNewAdapters_EmptyRemoteServer(t *testing.T) {
	logger := newTestLogger()
	adapters := NewAdapters("", logger)

	require.NotNil(t, adapters)
	assert.Nil(t, adapters.AuditEventAdapter, "audit adapter should be nil when remote server is empty")
}

func TestNewAdapters_WithDifferentURLs(t *testing.T) {
	logger := newTestLogger()

	urls := []string{
		"http://localhost:8080",
		"https://api.example.com/audit",
		"http://192.168.1.1:9090/events",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			adapters := NewAdapters(url, logger)
			require.NotNil(t, adapters)
			assert.NotNil(t, adapters.AuditEventAdapter)
		})
	}
}
