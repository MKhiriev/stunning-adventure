package handlers

import (
	"github.com/MKhiriev/stunning-adventure/models"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func (h *Handler) MetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	metrics := []string{"gauge", "counter"}

	// check if request method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// check if URL has 3 parts of metric
	url := r.URL
	urlParts := strings.Split(url.String(), "/")[2:]
	if len(urlParts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	metricType, metricName, metricValue := urlParts[0], urlParts[1], urlParts[2]

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
