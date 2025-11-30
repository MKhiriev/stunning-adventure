package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-chi/chi/v5"
)

// BatchUpdateMetricJSON
//
// Description:
//
//	Handles batch update of metrics in JSON format.
//	Validates each metric and saves them via metricsService.
//
// Input:
//
//	metrics JSON: `[{"id":"PollCount","type":"counter","delta":2},{"id":"FreeMemory","type":"gauge","value":184582144}]`
//
// Responses:
//
//   - 200 OK - all metrics updated successfully
//   - 400 Bad Request - invalid JSON or metric validation failed
//   - 500 Internal Server Error - service error while saving
func (h *Handler) BatchUpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var metricsFromBody []models.Metrics

	// Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricsFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("BatchUpdateMetricJSON was called!")

	// update all values + validation
	if err := h.metricsService.SaveAll(ctx, metricsFromBody); err != nil {
		switch {
		case errors.Is(err, validators.ErrEmptyID) || errors.Is(err, validators.ErrEmptyType) || errors.Is(err, validators.ErrNoValue) || errors.Is(err, validators.ErrInvalidType):
			h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("passed metric is not valid")
			http.Error(w, "passed metric is not valid", http.StatusBadRequest)
			return
		default:
			h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("error occurred during metric update")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateMetricJSON
//
// Description:
//
//	Handles update of a single metric in JSON format.
//	Validates the metric, saves via metricsService, and returns updated state.
//
// Input:
//
//	metric JSON: `{"id":"PollCount","type":"counter","delta":2}`
//
// Responses:
//
//   - 200 OK - updated metric returned in JSON
//   - 400 Bad Request - invalid metric or JSON
//   - 500 Internal Server Error - service error while saving
func (h *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var metricFromBody models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("func", "*Handler.UpdateMetricJSON").Interface("metric from body", metricFromBody).Msg("UpdateMetricJSON was called!")

	var err error
	// 3. Update metric's value based on it's type + validation
	if metricFromBody, err = h.metricsService.Save(ctx, &metricFromBody); err != nil {
		switch {
		case errors.Is(err, validators.ErrEmptyID) || errors.Is(err, validators.ErrEmptyType) || errors.Is(err, validators.ErrNoValue) || errors.Is(err, validators.ErrInvalidType):
			h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Any("metric", metricFromBody).Msg("passed metric is not valid")
			http.Error(w, "passed metric is not valid", http.StatusBadRequest)
			return
		default:
			h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during metric update")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	// 4. Set Content type to `application/json`
	w.Header().Set("Content-Type", "application/json")
	// 5. marshal in JSON saved metric
	savedMetricJSON, err := json.Marshal(metricFromBody)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during marshalling saved metric to JSON")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 6. return updated metric
	w.WriteHeader(http.StatusOK)
	w.Write(savedMetricJSON)
}

// GetMetricJSON
//
// Description:
//
//	Retrieves a metric by ID and type.
//	Validates parameters and fetches via metricsService.
//
// Input:
//
//	metric JSON without value: `{"id":"PollCount","type":"counter"}`
//
// Responses:
//
//   - 200 OK - found metric in JSON
//   - 400 Bad Request - invalid metric parameters
//   - 404 Not Found - metric not found
//   - 500 Internal Server Error - error marshalling JSON
func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var metric models.Metrics

	// 1. Get JSON from the body - handler level
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 3. Find metric in memory
	foundMetric, err := h.metricsService.Get(ctx, &metric)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Any("metric to find", metric).Msg("metric not found")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("metric not found"))
			//http.Error(w, "metric not found", http.StatusNotFound)
			return
		case errors.Is(err, validators.ErrEmptyID) || errors.Is(err, validators.ErrEmptyType) || errors.Is(err, validators.ErrInvalidType):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Any("metric to find", metric).Msg("metric type is not valid")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("passed metric is not valid"))
			//http.Error(w, "passed metric is not valid", http.StatusBadRequest)
			return
		}
	}

	// 4. Marshal
	foundMetricJSON, err := json.Marshal(foundMetric)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Msg("error occurred during marshalling metric from memory to JSON")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 5. Set header and status code
	w.WriteHeader(http.StatusOK)
	w.Write(foundMetricJSON)
}

// MetricHandler
//
// Description:
//
//	Creates or updates a metric via URL parameters.
//	Validates name, type, and value, then saves via metricsService.
//
// Input (URL params):
//
//   - metricName: "PollCount"
//   - metricType: "counter"
//   - metricValue: "123"
//
// Responses:
//
//   - 200 OK - metric created or updated successfully
//   - 400 Bad Request - invalid metric parameters or value
//   - 404 Not Found - empty metric name
//   - 500 Internal Server Error - error saving metric
func (h *Handler) MetricHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Add("Content-Type", "text/plain")

	metric := models.Metrics{
		ID:    chi.URLParam(r, "metricName"),
		MType: chi.URLParam(r, "metricType"),
	}
	metricValue := chi.URLParam(r, "metricValue")

	// check passed metric name and type
	if err := h.metricValidator.Validate(ctx, metric, validators.ID, validators.MType); err != nil {
		switch {
		case errors.Is(err, validators.ErrEmptyID):
			h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("metric name (id) is empty or not found")
			w.WriteHeader(http.StatusNotFound)
			return

		case errors.Is(err, validators.ErrInvalidType) || errors.Is(err, validators.ErrEmptyType):
			h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("metric type is not valid or empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// create new metric + validate validate metric value
	metric, err := models.NewMetric(metric.ID, metric.MType, metricValue)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("error during metric creation")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.metricsService.Save(ctx, &metric)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("error during saving metric")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetricValue
//
// Description:
//
//	Returns the value of a metric by name and type from URL parameters.
//	Fetches the full metric via metricsService and outputs only the value.
//
// Input (URL params):
//
//   - metricName: "PollCount"
//   - metricType: "counter"
//
// Responses:
//
//   - 200 OK - metric value in plain text
//   - 400 Bad Request - invalid metric type
//   - 404 Not Found - metric not found or empty name
func (h *Handler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Add("Content-Type", "text/plain")
	id := chi.URLParam(r, "metricName")
	mType := chi.URLParam(r, "metricType")

	// business logic + validation
	metric, err := h.metricsService.Get(ctx, &models.Metrics{ID: id, MType: mType})

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound) || errors.Is(err, validators.ErrEmptyID):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricValue").Msg("metric name (id) is empty or not found")
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, validators.ErrInvalidType) || errors.Is(err, validators.ErrEmptyType):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricValue").Msg("metric type is not valid or empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// if metric is present
	h.logger.Info().Caller().Str("func", "*Handler.GetMetricValue").Any("metric", metric).Msg("found metric")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.getValueFromMetric(metric)))
}

// GetAllMetrics
//
// Description:
//
//	Returns an HTML page with all metrics.
//	Retrieves all metrics via metricsService and renders via HTML templates.
//
// Input:
//
//	none (GET request without body)
//
// Responses:
//
//   - 200 OK - HTML page with metrics table
//   - 500 Internal Server Error - error rendering templates or service failure
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// TODO hide all HTML creation logic under new service
	html, err := template.ParseFiles("web/template/all-metrics.html", "web/template/metrics-list.html")
	if err != nil || html == nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetAllMetrics").Msg("error during parsing html templates")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	allMetrics, err := h.metricsService.GetAll(ctx)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetAllMetrics").Msg("error getting all metrics from storage")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	type HTMLMetric struct {
		ID    string
		MType string
		Value string
	}

	allHTMLMetrics := make([]HTMLMetric, len(allMetrics))
	for idx, metric := range allMetrics {
		allHTMLMetrics[idx] = HTMLMetric{ID: metric.ID, MType: metric.MType, Value: h.getValueFromMetric(metric)}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = html.Execute(w, allHTMLMetrics)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetAllMetrics").Msg("error during executing html templates")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getValueFromMetric(metric models.Metrics) string {
	if metric.MType == models.Counter && metric.Delta != nil {
		return strconv.Itoa(int(*metric.Delta))
	}
	if metric.MType == models.Gauge && metric.Value != nil {
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}
	return ""
}
