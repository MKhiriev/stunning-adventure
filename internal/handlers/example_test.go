package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

// ExampleBatchUpdateMetricJSON demonstrates batch updating metrics via JSON.
func ExampleHandler_BatchUpdateMetricJSON() {
	h := initExampleHandler()

	body := `[{"id":"PollCount","type":"counter","delta":2},{"id":"FreeMemory","type":"gauge","value":184582144}]`
	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.BatchUpdateMetricJSON(w, req)

	resp := w.Result()
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

// ExampleUpdateMetricJSON demonstrates updating a single metric via JSON.
func ExampleHandler_UpdateMetricJSON() {
	h := initExampleHandler()

	body := `{"id":"PollCount","type":"counter","delta":2}`
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.UpdateMetricJSON(w, req)

	resp := w.Result()
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

// ExampleGetMetricValue demonstrates retrieving a single metric value via URL parameters.
func ExampleHandler_GetMetricValue() {
	h := initExampleHandler()

	req := httptest.NewRequest(http.MethodGet, "/value/counter/PollCount", nil)
	w := httptest.NewRecorder()

	h.GetMetricValue(w, req)

	resp := w.Result()
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

// ExampleGetMetricJSON demonstrates retrieving a metric as JSON.
func ExampleHandler_GetMetricJSON() {
	h := initExampleHandler()

	body := `{"id":"PollCount","type":"counter"}`
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.GetMetricJSON(w, req)

	resp := w.Result()
	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

type mockMetricsService struct{}

func (m *mockMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	return nil
}

func (m *mockMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return metric, nil
}

func (m *mockMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	switch metric.MType {
	case "counter":
		d := int64(2)
		return models.Metrics{ID: metric.ID, MType: "counter", Delta: &d}, nil
	case "gauge":
		v := 184582144.0
		return models.Metrics{ID: metric.ID, MType: "gauge", Value: &v}, nil
	default:
		return models.Metrics{}, nil
	}
}

func (m *mockMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	d := int64(2)
	v := 184582144.0
	return []models.Metrics{
		{ID: "PollCount", MType: "counter", Delta: &d},
		{ID: "FreeMemory", MType: "gauge", Value: &v},
	}, nil
}

func initExampleHandler() *Handler {
	noLogs := zerolog.Nop()
	return &Handler{
		logger:          &noLogs, // no logs
		metricsService:  &mockMetricsService{},
		dbPingService:   nil, // not used in examples
		auditService:    nil, // disable audit
		metricValidator: validators.NewMetricsValidator(),
		hashKey:         "",
	}
}
