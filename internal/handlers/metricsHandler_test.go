package handlers

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
			res, _ := testRequest(t, ts, test.httpMethod, test.route, nil)
			defer res.Body.Close()

			// check if status code is correct
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestJSONGetMetricValue(t *testing.T) {
	h := initHandler()
	ts := httptest.NewServer(h.Init())
	defer ts.Close()

	type want struct {
		code        int
		response    string
		contentType string
		metric      models.Metrics
	}
	tests := []struct {
		name           string
		passedJSONBody string
		route          string
		httpMethod     string
		storedMetrics  map[string]models.Metrics
		want           want
	}{
		{
			name:           "positive counter test #1",
			passedJSONBody: `{"id": "someMetric", "type": "counter", "delta": 527}`,
			route:          "/update/",
			httpMethod:     http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				metric:      models.Metrics{ID: "someMetric", MType: "counter", Delta: mDelta(527)},
			},
		},
		{
			name:           "positive gauge test #2",
			passedJSONBody: `{"id": "Alloc", "type": "gauge", "value": 9999999}`,
			route:          "/update/",
			httpMethod:     http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				metric:      models.Metrics{ID: "Alloc", MType: "gauge", Value: mValue(9999999)},
			},
		},
		{
			name:           "negative test #3 - no metric's name",
			passedJSONBody: `{"type": "gauge", "value": 9999999}`,
			route:          "/update/",
			httpMethod:     http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name:           "negative test #4 - http.Method is not POST",
			passedJSONBody: `{"id": "Alloc", "type": "gauge", "value": 9999999}`,
			route:          "/update/",
			httpMethod:     http.MethodGet,
			want: want{
				code:        http.StatusNotFound,
				contentType: "",
			},
		},
		{
			name:           "negative test #5 - unknown type of metric",
			passedJSONBody: `{"id": "Alloc", "type": "otherType", "value": 9999999}`,
			route:          "/update/",
			httpMethod:     http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:           "negative test #6 - wrong value type",
			passedJSONBody: `{"id": "Alloc", "type": "gauge", "value": "wrongValue"}`,
			route:          "/update/",
			httpMethod:     http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, responseBody := testRequest(t, ts, test.httpMethod, test.route, strings.NewReader(test.passedJSONBody))
			defer response.Body.Close()

			// check HTTP status code and content type
			assert.Equal(t, test.want.code, response.StatusCode)
			assert.Equal(t, test.want.contentType, response.Header.Get("Content-Type"))

			// if status is 200 - check passed metric in http body
			if response.StatusCode == http.StatusOK {
				var receivedMetric models.Metrics
				jsonDecodingError := json.Unmarshal([]byte(responseBody), &receivedMetric)
				assert.NoError(t, jsonDecodingError)
				assert.True(t, areMetricsEqual(test.want.metric, receivedMetric))
			}
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
	return NewHandler(&logger)
}

func mDelta(v int) *int64 {
	deltaValue := int64(v)
	return &deltaValue
}

func mValue(v float64) *float64 {
	return &v
}

// testRequest from Yandex.Practicum
func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func areMetricsEqual(expected, actual models.Metrics) bool {
	if expected.MType == models.Gauge {
		return expected.ID == actual.ID && expected.MType == actual.MType && *expected.Value == *actual.Value
	} else {
		return expected.ID == actual.ID && expected.MType == actual.MType && *expected.Delta == *actual.Delta
	}

}
