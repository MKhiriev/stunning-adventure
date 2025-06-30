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
	h.Logger.Info().Msg("h.UpdateMetricJSON() was called")
	allowedMetricTypes := []string{models.Gauge, models.Counter}
	var metricFromBody models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBody); err != nil {
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 2. Check if metric from JSON is valid => if not - StatusBadRequest
	if metricFromBody.ID == "" || metricFromBody.MType == "" || !slices.Contains(allowedMetricTypes, metricFromBody.MType) {
		http.Error(w, "Passed metric is not valid", http.StatusBadRequest)
		return
	}

	// 3. Update metric's value based on it's type - first you need to do it ugly
	if metricFromBody.MType == models.Gauge {
		_, err := h.MemStorage.UpdateGauge(metricFromBody)
		if err != nil {
			http.Error(w, "error occured during gauge metric update", http.StatusInternalServerError)
			return
		}
	} else {
		_, err := h.MemStorage.AddCounter(metricFromBody)
		if err != nil {
			http.Error(w, "error occured during gauge metric update", http.StatusInternalServerError)
			return
		}
	}

	// 4. Set Content type to `application/json`
	w.Header().Set("Content-Type", "application/json")
	// 5. marshal in JSON saved metric
	savedMetricJSON, err := json.Marshal(metricFromBody)
	if err != nil {
		http.Error(w, "error occured during marshalling saved metric to JSON", http.StatusInternalServerError)
		return
	}

	// 6. return updated metric
	w.WriteHeader(http.StatusOK)
	w.Write(savedMetricJSON)
}

func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info().Msg("h.GetMetricJSON() was called")
	w.Header().Set("Content-Type", "application/json")
	allowedMetricTypes := []string{models.Gauge, models.Counter}
	var metricFromBodyWithoutValue models.Metrics

	// 1. Get JSON from the body
	if err := json.NewDecoder(r.Body).Decode(&metricFromBodyWithoutValue); err != nil {
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	// 2. Check if metric from JSON is valid => if not - StatusBadRequest
	if metricFromBodyWithoutValue.ID == "" || metricFromBodyWithoutValue.MType == "" || !slices.Contains(allowedMetricTypes, metricFromBodyWithoutValue.MType) {
		http.Error(w, "Passed metric is not valid", http.StatusBadRequest)
		return
	}

	// 3. Find metric in memory
	foundMetric, ok := h.MemStorage.GetMetricByNameAndType(metricFromBodyWithoutValue.ID, metricFromBodyWithoutValue.MType)

	if !ok {
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
		http.Error(w, "error occured during marshalling saved metric to JSON", http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// if metric type is not valued => http.StatusBadRequest
	if !slices.Contains(metrics, metricType) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	gaugeMetricToSave := models.Metrics{
		ID:    metricName,
		MType: models.Gauge,
		Value: &gaugeValue,
	}

	_, err := h.MemStorage.UpdateGauge(gaugeMetricToSave)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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

	_, err := h.MemStorage.AddCounter(counterMetricToSave)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	metrics := []string{models.Gauge, models.Counter}
	if !slices.Contains(metrics, metricType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, isMetricFound := h.MemStorage.GetMetricByNameAndType(metricName, metricType)

	// if metric is present
	if isMetricFound {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(
			h.getValueFromMetric(metric),
		))
	} else {
		// if not present - return not found
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	html, err := template.ParseFiles("web/template/all-metrics.html", "web/template/metrics-list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	allMetrics := h.MemStorage.GetAllMetrics()
	type HTMLMetric struct {
		ID    string
		MType string
		Value string
	}

	allHTMLMetrics := make([]HTMLMetric, len(allMetrics))
	for idx, metric := range allMetrics {
		allHTMLMetrics[idx] = HTMLMetric{ID: metric.ID, MType: metric.MType, Value: h.getValueFromMetric(metric)}
	}

	err = html.Execute(w, allHTMLMetrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
