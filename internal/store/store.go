package store

import (
	"errors"
	"github.com/MKhiriev/stunning-adventure/models"
	"maps"
	"slices"
)

type MetricsStorage interface {
	AddCounter(models.Metrics) (models.Metrics, error)
	UpdateGauge(models.Metrics) (models.Metrics, error)
	GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool)
	GetAllMetrics() []models.Metrics
}
type MemStorage struct {
	Memory map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Memory: make(map[string]models.Metrics)}
}

func (m *MemStorage) AddCounter(metric models.Metrics) (models.Metrics, error) {
	if metric.MType != models.Counter {
		return models.Metrics{}, errors.New("metric type is not `counter`")
	}

	val, ok := m.Memory[metric.ID]
	// if metric name exists in storage - apply Counter logic
	if ok {
		newDelta := *val.Delta + *metric.Delta
		val.Delta = &newDelta

		m.Memory[metric.ID] = val
		return val, nil
	}

	// if metric name doesn't exist - add it
	m.Memory[metric.ID] = metric
	return metric, nil
}

func (m *MemStorage) UpdateGauge(metric models.Metrics) (models.Metrics, error) {
	if metric.MType != models.Gauge {
		return models.Metrics{}, errors.New("metric type is not `gauge`")
	}

	val, ok := m.Memory[metric.ID]
	// if metric name exists in storage - apply Gauge logic
	if ok {
		val.Value = metric.Value
		m.Memory[metric.ID] = val
		return val, nil
	}

	// if metric name doesn't exist - add it
	m.Memory[metric.ID] = metric
	return metric, nil
}

func (m *MemStorage) GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool) {
	foundMetric, ok := m.Memory[metricName]
	if ok {
		if foundMetric.MType == metricType {
			return foundMetric, true
		}
		return models.Metrics{}, false
	}

	return emptyMetric(metricName, metricType), false
}

func (m *MemStorage) GetAllMetrics() []models.Metrics {
	return slices.Collect(maps.Values(m.Memory))
}

func emptyMetric(metricName string, metricType string) models.Metrics {
	if metricType == models.Gauge {
		return models.Metrics{ID: metricName, MType: metricType, Value: emptyValue()}
	} else {
		return models.Metrics{ID: metricName, MType: metricType, Delta: emptyDelta()}
	}
}

func emptyValue() *float64 {
	var v float64
	return &v
}

func emptyDelta() *int64 {
	var d int64
	return &d
}
