package agent

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadMetrics(t *testing.T) {
	agent := initAgent()

	type want struct {
		length int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				length: 29,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := agent.ReadMetrics()
			require.NoError(t, err)
			assert.NotEmpty(t, agent.memory.metrics)
			assert.Equal(t, test.want.length, len(agent.memory.metrics))
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
	agent := initAgent()

	type want struct {
		code          int
		response      string
		contentType   string
		route         string
		expectedDelta string
		expectedValue string
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
				code:          http.StatusOK,
				contentType:   "text/plain",
				route:         "/update/counter/someMetric/527",
				expectedDelta: "527",
			},
		},
		{
			name:       "positive gauge test #2",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(12779.105)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "text/plain",
				route:         "/update/gauge/someMetric/12779.105",
				expectedValue: "12779.105",
			},
		},
		{
			name:       "positive gauge test #3",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(575962.373)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "text/plain",
				route:         "/update/gauge/someMetric/575962.373",
				expectedValue: "575962.373",
			},
		},
		{
			name:       "positive gauge test #4",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(369111.063)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "text/plain",
				route:         "/update/gauge/someMetric/369111.063",
				expectedValue: "369111.063",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent.memory.metrics = map[string]models.Metrics{
				test.metric.ID: test.metric,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.URL.Path, test.want.route)
				assert.Contains(t, strings.Split(r.URL.Path, "/"), test.metric.ID)
				assert.Contains(t, strings.Split(r.URL.Path, "/"), test.metric.MType)
				if test.metric.MType == models.Counter {
					assert.Contains(t, strings.Split(r.URL.Path, "/"), test.want.expectedDelta)
				}
				if test.metric.MType == models.Gauge {
					assert.Contains(t, strings.Split(r.URL.Path, "/"), test.want.expectedValue)
				}

				assert.Equal(t, test.want.contentType, r.Header.Get("Content-Type"))

				w.Header().Add("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			agent.serverAddress = server.URL

			sendMetricsError := agent.SendMetrics()
			require.NoError(t, sendMetricsError)
		})
	}
}

func initAgent() *MetricsAgent {
	cfg := &config.AgentConfig{
		ServerAddress:  "0.0.0.0",
		ReportInterval: 2,
		PollInterval:   1,
	}
	return NewMetricsAgent("update", cfg, &zerolog.Logger{})
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v float64) *float64 {
	return &v
}
