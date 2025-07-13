package handlers

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
	"slices"
	"strconv"
)

func (h *Handler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	allowedMetricTypes := []string{models.Gauge, models.Counter}
	var metricFromBody models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBody); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 2. Check if metric from JSON is valid => if not - StatusBadRequest
	if metricFromBody.ID == "" || metricFromBody.MType == "" || !slices.Contains(allowedMetricTypes, metricFromBody.MType) {
		h.logger.Error().Caller().Str("func", "*Handler.UpdateMetricJSON").Any("metric", metricFromBody).Msg("passed metric is not valid")
		http.Error(w, "Passed metric is not valid", http.StatusBadRequest)
		return
	}

	var err error
	// 3. Update metric's value based on it's type - first you need to do it ugly
	if metricFromBody.MType == models.Gauge {
		metricFromBody, err = h.memStorage.UpdateGauge(metricFromBody)
		if err != nil {
			h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during gauge metric update")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		metricFromBody, err = h.memStorage.AddCounter(metricFromBody)
		if err != nil {
			h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during gauge metric update")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	if err := h.fileStorage.SaveMetricsToFile(h.memStorage.GetAllMetrics()); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.UpdateMetricJSON").Msg("error occurred during metrics save to file")
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
	allowedMetricTypes := []string{models.Gauge, models.Counter}
	var metricFromBodyWithoutValue models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBodyWithoutValue); err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 2. Check if metric from JSON is valid => if not - StatusBadRequest
	if metricFromBodyWithoutValue.ID == "" || metricFromBodyWithoutValue.MType == "" || !slices.Contains(allowedMetricTypes, metricFromBodyWithoutValue.MType) {
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricJSON").Any("metric", metricFromBodyWithoutValue).Msg("passed metric is not valid")
		http.Error(w, "passed metric is not valid", http.StatusBadRequest)
		return
	}

	// 3. Find metric in memory
	foundMetric, ok := h.memStorage.GetMetricByNameAndType(metricFromBodyWithoutValue.ID, metricFromBodyWithoutValue.MType)

	if !ok {
		h.logger.Info().Caller().Str("func", "*Handler.GetMetricJSON").Any("metric to find", metricFromBodyWithoutValue).Msg("metric not found")
		w.WriteHeader(http.StatusNotFound)
	}
	if metricFromBodyWithoutValue.MType == models.Gauge && foundMetric.Value != nil {
		metricFromBodyWithoutValue.Value = foundMetric.Value
		w.WriteHeader(http.StatusOK)
	}
	if metricFromBodyWithoutValue.MType == models.Counter && foundMetric.Delta != nil {
		metricFromBodyWithoutValue.Delta = foundMetric.Delta
		w.WriteHeader(http.StatusOK)
	}

	// 4. Marshal
	foundMetricJSON, err := json.Marshal(metricFromBodyWithoutValue)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetMetricJSON").Msg("error occurred during marshalling metric from memory to JSON")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 5. Set header and status code
	w.Write(foundMetricJSON)
}

func (h *Handler) MetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	metrics := []string{models.Gauge, models.Counter}

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	// if metric name is not specified
	if metricName == "" {
		h.logger.Error().Caller().Str("func", "*Handler.MetricHandler").Msg("metric name is not specified")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// if metric type is not valid => http.StatusBadRequest
	if !slices.Contains(metrics, metricType) {
		h.logger.Error().Caller().Str("func", "*Handler.MetricHandler").Msg("metric type is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if metricType == models.Counter {
		h.HandleCounter(w, metricValue, metricName)
		return
	}
	if metricType == models.Gauge {
		h.HandleGauge(w, metricValue, metricName)
		return
	}
}

func (h *Handler) HandleGauge(w http.ResponseWriter, metricValue string, metricName string) {
	gaugeValue, conversionError := strconv.ParseFloat(metricValue, 64)
	// if metric value type is wrong => http.StatusBadRequest
	if conversionError != nil {
		h.logger.Error().Caller().Str("func", "*Handler.HandleGauge").Msg("metric value type is wrong")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	gaugeMetricToSave := models.Metrics{
		ID:    metricName,
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	_, err := h.memStorage.UpdateGauge(gaugeMetricToSave)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.HandleGauge").Msg("error during updating gauge metric")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.fileStorage.SaveMetricsToFile(h.memStorage.GetAllMetrics())
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.HandleGauge").Msg("error during saving all metrics to file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleCounter(w http.ResponseWriter, metricValue string, metricName string) {
	counterValue, conversionError := strconv.ParseInt(metricValue, 10, 64)
	// if metric value type is wrong => http.StatusBadRequest
	if conversionError != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	counterMetricToSave := models.Metrics{
		ID:    metricName,
		MType: models.Counter,
		Delta: &counterValue,
	}

	_, err := h.memStorage.AddCounter(counterMetricToSave)
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.HandleCounter").Msg("error during updating counter metric")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.fileStorage.SaveMetricsToFile(h.memStorage.GetAllMetrics())
	if err != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.HandleCounter").Msg("error during saving all metrics to file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricName == "" {
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricValue").Msg("metric name is not specified")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metrics := []string{models.Gauge, models.Counter}
	if !slices.Contains(metrics, metricType) {
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricValue").Msg("metric type is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, isMetricFound := h.memStorage.GetMetricByNameAndType(metricName, metricType)

	// if metric is present
	if isMetricFound {
		h.logger.Info().Caller().Str("func", "*Handler.GetMetricValue").Any("metric", metric).Msg("found metric")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(
			h.getValueFromMetric(metric),
		))
	} else {
		// if not present - return not found
		h.logger.Error().Caller().Str("func", "*Handler.GetMetricValue").Msg("metric not found")
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web/template/all-metrics.html", "web/template/metrics-list.html")
	if err != nil || html != nil {
		h.logger.Err(err).Caller().Str("func", "*Handler.GetAllMetrics").Msg("error during parsing html templates")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	allMetrics := h.memStorage.GetAllMetrics()
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
