package agent

import (
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		code        int
		response    string
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
				contentType: "text/plain",
				route:       "/update/counter/someMetric/527",
			},
		},
		{
			name:       "positive gauge test #2",
			metric:     models.Metrics{ID: "someMetric", MType: models.Gauge, Value: mValue(527)},
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				route:       "/update/gauge/someMetric/527",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent.Memory.Metrics = map[string]models.Metrics{
				test.metric.ID: test.metric,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.URL.Path, test.want.route)
				assert.Contains(t, strings.Split(r.URL.Path, "/"), test.metric.ID)
				assert.Contains(t, strings.Split(r.URL.Path, "/"), test.metric.MType)
				if test.metric.MType == models.Counter {
					assert.Contains(t, strings.Split(r.URL.Path, "/"), strconv.Itoa(int(*test.metric.Delta)))
				}
				if test.metric.MType == models.Gauge {
					assert.Contains(t, strings.Split(r.URL.Path, "/"), strconv.FormatFloat(*test.metric.Value, 'f', -1, 64))
				}

				assert.Equal(t, test.want.contentType, r.Header.Get("Content-Type"))

				w.Header().Add("Content-Type", "text/plain")
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
	return NewMetricsAgent("0.0.0.0", "update")
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v int) *float64 {
	value := float64(v)
	return &value
}
