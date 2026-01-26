package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMetrics(t *testing.T) {
	agent := initAgent(t, &config.AgentConfig{
		ServerAddress:  "localhost:8099",
		ReportInterval: 2,
		PollInterval:   1,
	})

	type want struct {
		moreThanZeroMetrics bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				moreThanZeroMetrics: len(agent.memory.metrics) > 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := agent.ReadMetrics()
			require.NoError(t, err)
			assert.NotEmpty(t, agent.memory.metrics)
			assert.True(t, len(agent.memory.metrics) > 0)
			// check for non nil values
			for _, metric := range agent.memory.metrics {
				if metric.MType == models.Gauge {
					assert.NotNil(t, metric.Value)
				}
				if metric.MType == models.Counter {
					assert.NotNil(t, metric.Delta)
				}
			}
		})
	}
}

func TestSendMetrics(t *testing.T) {
	type want struct {
		code        int
		contentType string
		route       string
	}
	tests := []struct {
		name       string
		metric     models.Metrics
		httpMethod string
		want       want
	}{
		{
			name:       "positive counter test #1",
			metric:     models.Metrics{ID: "someMetric", MType: models.Counter, Delta: mDelta(527)},
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				route:       "/update",
			},
		},
		{
			name:       "positive gauge test #2",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(12779.105)},
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				route:       "/update",
			},
		},
		{
			name:       "positive gauge test #3",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(575962.373)},
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				route:       "/update",
			},
		},
		{
			name:       "positive gauge test #4",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(369111.063)},
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				route:       "/update",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, test.httpMethod, r.Method)
				assert.Equal(t, test.want.route, r.URL.Path)
				assert.Equal(t, test.want.contentType, r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

				gzReader, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				defer gzReader.Close()

				body, err := io.ReadAll(gzReader)
				require.NoError(t, err)

				var metrics []models.Metrics
				err = json.Unmarshal(body, &metrics)
				require.NoError(t, err)
				require.Len(t, metrics, 1)
				assert.Equal(t, test.metric, metrics[0])

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(models.Metrics{})
			}))
			defer server.Close()

			cfg := &config.AgentConfig{
				ServerAddress:  utils.ServerAddress(server.URL),
				ReportInterval: 2,
				PollInterval:   1,
			}
			agent := initAgent(t, cfg)

			agent.memory.metrics = map[string]models.Metrics{
				test.metric.ID: test.metric,
			}

			metrics := agent.memory.GetAllMetrics()
			sendMetricsError := agent.sendMetrics(metrics...)
			require.NoError(t, sendMetricsError)
		})
	}
}

func initAgent(t *testing.T, cfg *config.AgentConfig) *MetricsAgent {
	logger := zerolog.Nop()
	agent, err := NewMetricsAgent("update", nil, cfg, &logger)
	require.NoError(t, err, "failed to create metrics agent")

	return agent
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v float64) *float64 {
	return &v
}
