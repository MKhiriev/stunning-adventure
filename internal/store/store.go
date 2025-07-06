package store

import (
	"errors"
	"github.com/MKhiriev/stunning-adventure/models"
	"maps"
	"slices"
	"sync"
)

type MetricsStorage interface {
	AddCounter(models.Metrics) (models.Metrics, error)
	UpdateGauge(models.Metrics) (models.Metrics, error)
	GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool)
	GetAllMetrics() []models.Metrics
}
type MemStorage struct {
	Memory map[string]models.Metrics `json:"metrics"`
	mu     sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Memory: make(map[string]models.Metrics)}
}

func (m *MemStorage) AddCounter(metrics models.Metrics) (models.Metrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics.MType != models.Counter {
		return models.Metrics{}, errors.New("metric type is not `counter`")
	}

	val, ok := m.Memory[metrics.ID]
	// if metric name exists in storage - apply Counter logic
	if ok {
		newDelta := *val.Delta + *metrics.Delta
		val.Delta = &newDelta

		m.Memory[metrics.ID] = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metrics.ID] = metrics
	}

	return models.Metrics{}, nil
}

func (m *MemStorage) UpdateGauge(metrics models.Metrics) (models.Metrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics.MType != models.Gauge {
		return models.Metrics{}, errors.New("metric type is not `gauge`")
	}

	val, ok := m.Memory[metrics.ID]
	// if metric name exists in storage - apply Gauge logic
	if ok {
		val.Value = metrics.Value
		m.Memory[metrics.ID] = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metrics.ID] = metrics
	}

	return models.Metrics{}, nil
}

func (m *MemStorage) GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	foundMetric, ok := m.Memory[metricName]
	if ok {
		if foundMetric.MType == metricType {
			return foundMetric, true
		}
		return models.Metrics{}, false
	}

	return models.Metrics{}, false
}

func (m *MemStorage) GetAllMetrics() []models.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return slices.Collect(maps.Values(m.Memory))
}
