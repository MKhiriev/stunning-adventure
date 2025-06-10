package handlers

import (
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
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
		return
	}
	if metricType == models.Gauge {
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
		return
	}
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
	if isMetricFound == true {
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
		log.Println("GetAllMetrics(): error occurred during reading html file")
		return
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
}

func (h *Handler) getValueFromMetric(metric models.Metrics) string {
	if metric.MType == models.Counter && metric.Delta != nil {
		return strconv.Itoa(int(*metric.Delta))
	}
	if metric.MType == models.Gauge && metric.Value != nil {
		return strconv.FormatFloat(*metric.Value, 'f', 0, 64)
	}
	return ""
}
