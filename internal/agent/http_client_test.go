package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPMetricsSender(t *testing.T) {
	logger := zerolog.Nop()
	serverAddress := "127.0.0.1:8081"
	route := "updates"
	retryIntervals := map[int]time.Duration{
		1: 1 * time.Second,
		2: 3 * time.Second,
		3: 5 * time.Second,
	}

	sender, err := NewHTTPMetricsSender(serverAddress, route, retryIntervals, nil, nil, &logger)

	require.NoError(t, err)
	assert.NotNil(t, sender)
	assert.Equal(t, route, sender.route)
	assert.NotNil(t, sender.client)
	assert.NotEmpty(t, sender.realIP)
	assert.Equal(t, &logger, sender.logger)
}

func TestHTTPMetricsSender_sendMetrics_NoMetrics(t *testing.T) {
	logger := zerolog.Nop()
	sender := &HTTPMetricsSender{
		logger: &logger,
	}

	err := sender.sendMetrics()
	assert.NotNil(t, err)
	assert.Error(t, err)
	assert.Equal(t, "no metric was passed", err.Error())
}

func TestHTTPMetricsSender_sendMetrics_Success(t *testing.T) {
	logger := zerolog.Nop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/update/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.NotEmpty(t, r.Header.Get("X-Real-IP"))

		gzReader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		require.NoError(t, err)

		var metrics []models.Metrics
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)
		assert.Len(t, metrics, 1)
		assert.Equal(t, "test_metric", metrics[0].ID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Metrics{})
	}))
	defer server.Close()

	client := utils.NewHTTPClient(5 * time.Second)
	client.SetBaseURL(server.URL)

	gaugeValue := 123.45
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	sender := &HTTPMetricsSender{
		route:  "/update/",
		client: client,
		logger: &logger,
		realIP: "127.0.0.1",
	}

	err := sender.sendMetrics(metric)
	assert.NoError(t, err)
}

func TestHTTPMetricsSender_sendMetrics_WithHasher(t *testing.T) {
	logger := zerolog.Nop()
	hasher := utils.NewHasher("test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("HashSHA256"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Metrics{})
	}))
	defer server.Close()

	client := utils.NewHTTPClient(5 * time.Second)
	client.SetBaseURL(server.URL)

	counterValue := int64(42)
	metric := models.Metrics{
		ID:    "test_counter",
		MType: models.Counter,
		Delta: &counterValue,
	}

	sender := &HTTPMetricsSender{
		route:  "/update/",
		client: client,
		hasher: hasher,
		logger: &logger,
		realIP: "127.0.0.1",
	}

	err := sender.sendMetrics(metric)
	assert.NoError(t, err)
}

func TestHTTPMetricsSender_sendMetrics_WithEncryption(t *testing.T) {
	logger := zerolog.Nop()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Metrics{})
	}))
	defer server.Close()

	client := utils.NewHTTPClient(5 * time.Second)
	client.SetBaseURL(server.URL)

	gaugeValue := 99.99
	metric := models.Metrics{
		ID:    "encrypted_metric",
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	sender := &HTTPMetricsSender{
		route:     "/update/",
		client:    client,
		publicKey: publicKey,
		logger:    &logger,
		realIP:    "127.0.0.1",
	}

	err = sender.sendMetrics(metric)
	assert.NoError(t, err)
}

func TestHTTPMetricsSender_sendMetrics_MultipleMetrics(t *testing.T) {
	logger := zerolog.Nop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzReader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		require.NoError(t, err)

		var metrics []models.Metrics
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)
		assert.Len(t, metrics, 3)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Metrics{})
	}))
	defer server.Close()

	client := utils.NewHTTPClient(5 * time.Second)
	client.SetBaseURL(server.URL)

	gaugeValue1 := 1.1
	gaugeValue2 := 2.2
	counterValue := int64(10)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: &gaugeValue1},
		{ID: "gauge2", MType: models.Gauge, Value: &gaugeValue2},
		{ID: "counter1", MType: models.Counter, Delta: &counterValue},
	}

	sender := &HTTPMetricsSender{
		route:  "/update/",
		client: client,
		logger: &logger,
		realIP: "127.0.0.1",
	}

	err := sender.sendMetrics(metrics...)
	assert.NoError(t, err)
}

func TestHTTPMetricsSender_sendMetrics_WithHasherAndEncryption(t *testing.T) {
	logger := zerolog.Nop()
	hasher := utils.NewHasher("test-key")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("HashSHA256"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Metrics{})
	}))
	defer server.Close()

	client := utils.NewHTTPClient(5 * time.Second)
	client.SetBaseURL(server.URL)

	gaugeValue := 55.55
	counterValue := int64(100)

	metrics := []models.Metrics{
		{ID: "gauge_encrypted", MType: models.Gauge, Value: &gaugeValue},
		{ID: "counter_encrypted", MType: models.Counter, Delta: &counterValue},
	}

	sender := &HTTPMetricsSender{
		route:     "/update/",
		client:    client,
		hasher:    hasher,
		publicKey: publicKey,
		logger:    &logger,
		realIP:    "192.168.1.1",
	}

	err = sender.sendMetrics(metrics...)
	assert.NoError(t, err)
}

func Test_gzipCompress(t *testing.T) {
	tests := []struct {
		name    string
		metrics []models.Metrics
		wantErr bool
	}{
		{
			name:    "no metrics",
			metrics: []models.Metrics{},
			wantErr: true,
		},
		{
			name: "single metric",
			metrics: []models.Metrics{
				{ID: "test", MType: models.Gauge, Value: func() *float64 { v := 1.1; return &v }()},
			},
			wantErr: false,
		},
		{
			name: "multiple metrics",
			metrics: []models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: func() *float64 { v := 1.1; return &v }()},
				{ID: "test2", MType: models.Counter, Delta: func() *int64 { v := int64(10); return &v }()},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gzipCompress(tt.metrics...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, got)

			gzReader, err := gzip.NewReader(bytes.NewReader(got))
			require.NoError(t, err)
			defer gzReader.Close()

			decompressed, err := io.ReadAll(gzReader)
			require.NoError(t, err)

			var decodedMetrics []models.Metrics
			err = json.Unmarshal(decompressed, &decodedMetrics)
			require.NoError(t, err)

			assert.Equal(t, tt.metrics, decodedMetrics)
		})
	}
}
