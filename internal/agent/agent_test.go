package agent

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
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
			assert.NotEmpty(t, agent.Memory.Metrics)
			assert.Equal(t, test.want.length, len(agent.Memory.Metrics))
			// check for non nil values
			for _, metric := range agent.Memory.Metrics {
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
		expectedID    string
		expectedDelta int64
		expectedValue float64
		expectedMType string
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
				contentType:   "application/json",
				route:         "/update/",
				expectedID:    "someMetric",
				expectedDelta: 527,
				expectedMType: models.Counter,
			},
		},
		{
			name:       "positive gauge test #2",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(12779.105)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "application/json",
				route:         "/update/",
				expectedID:    "someMetric",
				expectedValue: 12779.105,
				expectedMType: models.Gauge,
			},
		},
		{
			name:       "positive gauge test #3",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(575962.373)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "application/json",
				route:         "/update/",
				expectedID:    "someMetric",
				expectedValue: 575962.373,
				expectedMType: models.Gauge,
			},
		},
		{
			name:       "positive gauge test #4",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(369111.063)},
			httpMethod: http.MethodPost,
			want: want{
				code:          http.StatusOK,
				contentType:   "application/json",
				route:         "/update/",
				expectedID:    "someMetric",
				expectedValue: 369111.063,
				expectedMType: models.Gauge,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent.Memory.Metrics = map[string]models.Metrics{
				test.metric.ID: test.metric,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.URL.Path)               // check if URL Path is not ""
				assert.Equal(t, r.URL.Path, test.want.route) // check if URL Path has desired value

				// extract from metric from body
				var receivedMetric models.Metrics
				jsonDecodingError := json.NewDecoder(r.Body).Decode(&receivedMetric)
				assert.NoError(t, jsonDecodingError) // check if there are no error during getting metric from JSON in HTTP Body

				assert.Equal(t, test.want.expectedID, receivedMetric.ID)       // check if ID is correct
				assert.Equal(t, test.want.expectedMType, receivedMetric.MType) // check if MType is correct

				// check if metric value is correct
				if test.metric.MType == models.Counter {
					assert.Equal(t, test.want.expectedDelta, *receivedMetric.Delta)
				}
				if test.metric.MType == models.Gauge {
					assert.Equal(t, test.want.expectedValue, *receivedMetric.Value)
				}

				// check if `Content-Type` is correct
				assert.Equal(t, test.want.contentType, r.Header.Get("Content-Type"))

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			agent.ServerAddress = server.URL

			sendMetricsError := agent.SendMetrics()
			require.NoError(t, sendMetricsError)
		})
	}
}

func initAgent() *MetricsAgent {
	return NewMetricsAgent("update/", &config.AgentConfig{
		ServerAddress:  "0.0.0.0",
		ReportInterval: 2,
		PollInterval:   1,
	})
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v float64) *float64 {
	return &v
}
