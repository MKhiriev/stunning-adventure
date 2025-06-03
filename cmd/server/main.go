package main

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

var memory map[string]models.Metrics

func main() {
	memory = store.MemStorage{Memory: make(map[string]models.Metrics)}.Memory
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", MetricHandler) // {metric_type}/{metric_name}/{metric_value}
	err := http.ListenAndServe("localhost:8080", mux)
	log.Fatal(err)
}

func MetricHandler(w http.ResponseWriter, r *http.Request) {
	metrics := []string{"gauge", "counter"}

	// check if request method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

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

	var metric models.Metrics
	metric.Delta = new(int64)
	metric.Value = new(float64)
	if metricType == models.Counter {
		counterValue, conversionError := strconv.ParseInt(metricValue, 10, 64)
		// if metric value type is wrong => http.StatusBadRequest
		if conversionError != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// define metric value
		metric.MType = models.Counter
		metric.Delta = &counterValue

		val, ok := memory[metricName]
		if ok {
			if val.Delta == nil {
				val.Delta = metric.Delta
			} else {
				newDelta := *val.Delta + *metric.Delta
				val.Delta = &newDelta
			}
			memory[metricName] = val
		} else {
			memory[metricName] = metric
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
		metric.MType = models.Gauge
		metric.Value = &gaugeValue

		val, ok := memory[metricName]
		if ok && val.Value != nil {
			val.Value = metric.Value
			memory[metricName] = val
		} else {
			memory[metricName] = metric
		}
		w.WriteHeader(http.StatusOK)
		return
	}
}
