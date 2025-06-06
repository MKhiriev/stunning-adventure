package store

import (
	"errors"
	"github.com/MKhiriev/stunning-adventure/models"
)

type MetricsStorage interface {
	AddCounter(models.Metrics) (models.Metrics, error)
	UpdateGauge(models.Metrics) (models.Metrics, error)
}
type MemStorage struct {
	Memory map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Memory: make(map[string]models.Metrics)}
}

func (m *MemStorage) AddCounter(metrics models.Metrics) (models.Metrics, error) {
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
