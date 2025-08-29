package handlers

import (
	"context"
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

func (h *Handler) BatchUpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metricsFromBody []models.Metrics

	// Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricsFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("func", "*Handler.BatchUpdateMetricJSON").Interface("metric from body", metricsFromBody).Msg("BatchUpdateMetricJSON was called!")

	for _, metric := range metricsFromBody {
		// Check if metric from JSON is valid => if not - StatusBadRequest
		if err := h.metricValidator.Validate(context.TODO(), metric); err != nil {
			h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Any("metric", metric).Msg("passed metric is not valid")
			http.Error(w, "passed metric is not valid", http.StatusBadRequest)
			return
		}
	}

	// update all values
	if err := h.metricsService.SaveAll(context.TODO(), metricsFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.BatchUpdateMetricJSON").Msg("error occurred during metric update")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 4. Set Content type to `application/json`
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metricFromBody models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("func", "*Handler.UpdateMetricJSON").Interface("metric from body", metricFromBody).Msg("UpdateMetricJSON was called!")

	// 2. Check if metric from JSON is valid => if not - StatusBadRequest
	if err := h.metricValidator.Validate(context.TODO(), metricFromBody); err != nil {
		h.logger.Error().Caller().Str("func", "*Handler.UpdateMetricJSON").Any("metric", metricFromBody).Msg("passed metric is not valid")
		http.Error(w, "passed metric is not valid", http.StatusBadRequest)
		return
	}

	var err error
	// 3. Update metric's value based on it's type
	metricFromBody, err = h.metricsService.Save(context.TODO(), metricFromBody)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during metric update")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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

func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var metric models.Metrics

	// 1. Get JSON from the body - handler level
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 2. Check if metric from JSON has non-empty fields => if not - StatusBadRequest
	if h.checkValue(metric.ID != "", validators.ErrEmptyID) != nil || h.checkValue(metric.MType != "", validators.ErrEmptyType) != nil {
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricJSON").Any("metric", metric).Msg("passed metric is not valid")
		http.Error(w, "passed metric is not valid", http.StatusBadRequest)
		return
	}

	// 3. Find metric in memory
	foundMetric, err := h.metricsService.Get(context.TODO(), metric)

	if err != nil {
		switch {
		case errors.Is(err, validators.ErrInvalidType):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Any("metric to find", metric).Msg("metric type is not valid")
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, store.ErrNotFound):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Any("metric to find", metric).Msg("metric not found")
			w.WriteHeader(http.StatusNotFound)
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

func (h *Handler) MetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	metric := models.Metrics{
		ID:    chi.URLParam(r, "metricName"),
		MType: chi.URLParam(r, "metricType"),
	}
	metricValue := chi.URLParam(r, "metricValue")

	// if metric name is not specified - iter1
	if err := h.metricValidator.Validate(context.TODO(), metric, validators.ID); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("metric name is not specified")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// if metric type is not valid => http.StatusBadRequest
	if err := h.metricValidator.Validate(context.TODO(), metric, validators.MType); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("metric type is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, err := models.NewMetric(metric.ID, metric.MType, metricValue)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("error during metric creation")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = h.metricsService.Save(context.TODO(), metric)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.MetricHandler").Msg("error during saving metric")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	id := chi.URLParam(r, "metricName")
	mType := chi.URLParam(r, "metricType")

	// check if empty url params - handler
	if err := h.checkValue(id != "", validators.ErrEmptyID); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricValue").Msg("metric name is not specified")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.checkValue(mType != "", validators.ErrEmptyType); err != nil {
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricValue").Msg("metric type is not specified")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// business logic + validation
	metric, err := h.metricsService.Get(context.TODO(), models.Metrics{ID: id, MType: mType})

	if err != nil {
		switch {
		case errors.Is(err, validators.ErrInvalidType):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricValue").Msg("metric type is not valid")
			w.WriteHeader(http.StatusBadRequest)
			return
		case errors.Is(err, store.ErrNotFound):
			h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricValue").Msg("metric not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	// if metric is present
	h.logger.Info().Caller().Str("func", "*Handler.GetMetricValue").Any("metric", metric).Msg("found metric")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.getValueFromMetric(metric)))
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web/template/all-metrics.html", "web/template/metrics-list.html")
	if err != nil || html == nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetAllMetrics").Msg("error during parsing html templates")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	allMetrics, err := h.metricsService.GetAll(context.TODO())
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

func (h *Handler) checkValue(cond bool, ifNotEmptyError error) error {
	if !cond {
		h.logger.Error().Str("func", "*Handler.checkValue").Msg("value is not valid")
		return ifNotEmptyError
	}

	return nil
}
