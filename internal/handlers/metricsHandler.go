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

func (h *Handler) JSONMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	metrics := []string{models.Gauge, models.Counter}

	metric := models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if metric name is not specified
	if metric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// if metric type is not valued => http.StatusBadRequest
	if !slices.Contains(metrics, metric.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.SaveMetric(w, metric)
}

func (h *Handler) SaveMetric(w http.ResponseWriter, metric models.Metrics) {
	var err error
	var savedMetric models.Metrics

	// save metric
	if metric.MType == models.Counter {
		savedMetric, err = h.MemStorage.AddCounter(metric)
	}
	if metric.MType == models.Gauge {
		savedMetric, err = h.MemStorage.UpdateGauge(metric)
	}

	// if error occurred during saving - return error with 500 code status
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// convert saved metric to JSON
	var savedMetricJSON []byte
	savedMetricJSON, err = json.Marshal(&savedMetric)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// return saved metric
	w.WriteHeader(http.StatusOK)
	w.Write(savedMetricJSON)
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

func (h *Handler) JSONGetMetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// get models.Metrics from HTTP Body
	requestedMetric := models.Metrics{}
	if err := json.NewDecoder(r.Body).Decode(&requestedMetric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if requestedMetric.ID == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	metrics := []string{models.Gauge, models.Counter}
	if !slices.Contains(metrics, requestedMetric.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric, isMetricFound := h.MemStorage.GetMetricByNameAndType(requestedMetric.ID, requestedMetric.MType)

	// convert saved metric to JSON
	var savedMetricJSON []byte
	savedMetricJSON, err := json.Marshal(&metric)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// if metric is present
	if isMetricFound {
		// return saved metric
		w.WriteHeader(http.StatusOK)
		w.Write(savedMetricJSON)
	} else {
		// if not present - return not found and empty metric
		w.WriteHeader(http.StatusNotFound)
		w.Write(savedMetricJSON)
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

func (h *Handler) getValueFromMetric(metric models.Metrics) string {
	if metric.MType == models.Counter && metric.Delta != nil {
		return strconv.Itoa(int(*metric.Delta))
	}
	if metric.MType == models.Gauge && metric.Value != nil {
		return strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}
	return ""
}
