package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMetricHandler(t *testing.T) {
	h := initHandler()
	ts := httptest.NewServer(h.Init())
	defer ts.Close()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name          string
		route         string
		httpMethod    string
		storedMetrics map[string]models.Metrics
		want          want
	}{
		{
			name:       "positive counter test #1",
			route:      "/update/counter/someMetric/527",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:       "positive gauge test #2",
			route:      "/update/gauge/Alloc/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #3 - no metric's name",
			route:      "/update/gauge/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:       "negative test #4 - http.Method is not POST",
			route:      "/update/gauge/Alloc/9999999",
			httpMethod: http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name:       "negative test #5 - unknown type of metric",
			route:      "/update/otherType/Alloc/9999999",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
		{
			name:       "negative test #6 - wrong value type",
			route:      "/update/otherType/Alloc/wrongValue",
			httpMethod: http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, _ := testRequest(t, ts, test.httpMethod, test.route)
			defer res.Body.Close()

			// check if status code is correct
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGetValueFromMetric(t *testing.T) {
	h := initHandler()
	type want struct {
		result string
	}
	tests := []struct {
		name   string
		metric models.Metrics
		want   want
	}{
		{
			name:   "positive gauge value test #1",
			metric: models.Metrics{ID: "Alloc", MType: models.Gauge, Value: mValue(10.0)},
			want:   want{result: "10"},
		},
		{
			name:   "positive gauge value test #2",
			metric: models.Metrics{ID: "Alloc", MType: models.Gauge, Value: mValue(123.229)},
			want:   want{result: "123.229"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val := h.getValueFromMetric(test.metric)
			assert.NotEmpty(t, val)
			assert.Equal(t, test.want.result, val)
		})
	}
}

func initHandler() *Handler {
	logger := zerolog.New(os.Stdout).With().Logger()
	cfg := &config.ServerConfig{
		ServerAddress:          "localhost:8080",
		StoreInterval:          300,
		FileStoragePath:        "internal/store/metrics.log",
		RestoreMetricsFromFile: false,
	}
	memStorage := store.NewMemStorage()
	fileStorage := store.NewFileStorage(memStorage, cfg)

	return NewHandler(memStorage, fileStorage, &logger)
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v float64) *float64 {
	return &v
}

// testRequest from Yandex.Practicum
func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
