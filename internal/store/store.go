package store

import (
	"context"
	"errors"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
	"maps"
	"slices"
	"sync"
)

type MemStorage struct {
	Memory map[string]models.Metrics `json:"metrics"`
	mu     *sync.Mutex
	log    *zerolog.Logger
}

func NewMemStorage(log *zerolog.Logger) *MemStorage {
	return &MemStorage{Memory: make(map[string]models.Metrics), mu: &sync.Mutex{}, log: log}
}

func (m *MemStorage) AddCounter(ctx context.Context, metrics models.Metrics) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result models.Metrics

	if metrics.MType != models.Counter {
		return models.Metrics{}, errors.New("metric type is not `counter`")
	}

	val, ok := m.Memory[metrics.ID]
	// if metric name exists in storage - apply Counter logic
	if ok {
		newDelta := *val.Delta + *metrics.Delta
		val.Delta = &newDelta

		m.Memory[metrics.ID] = val
		result = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metrics.ID] = metrics
		result = metrics
	}

	return result, nil
}

func (m *MemStorage) UpdateGauge(ctx context.Context, metrics models.Metrics) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result models.Metrics

	if metrics.MType != models.Gauge {
		return models.Metrics{}, errors.New("metric type is not `gauge`")
	}

	val, ok := m.Memory[metrics.ID]
	// if metric name exists in storage - apply Gauge logic
	if ok {
		val.Value = metrics.Value
		m.Memory[metrics.ID] = val
		result = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metrics.ID] = metrics
		result = metrics
	}

	return result, nil
}

func (m *MemStorage) GetMetricByNameAndType(ctx context.Context, metricName string, metricType string) (models.Metrics, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	foundMetric, ok := m.Memory[metricName]
	if ok {
		if foundMetric.MType == metricType {
			return foundMetric, true
		}
		return models.Metrics{}, false
	}

	return models.Metrics{}, false
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slices.Collect(maps.Values(m.Memory))
}

func (m *MemStorage) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	switch metric.MType {
	case models.Counter:
		return m.AddCounter(ctx, metric)
	case models.Gauge:
		return m.UpdateGauge(ctx, metric)
	default:
		return metric, errors.New("unsupported metric type")
	}
}

func (m *MemStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	var err error
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			_, err = m.AddCounter(ctx, metric)
			if err != nil {
				return err
			}
		case models.Gauge:
			_, err = m.UpdateGauge(ctx, metric)
			if err != nil {
				return err
			}
		default:
			return errors.New("unsupported metric type")
		}
	}
	return nil
}

func (m *MemStorage) Get(ctx context.Context, metric models.Metrics) (models.Metrics, bool) {
	return m.GetMetricByNameAndType(ctx, metric.ID, metric.MType)
}

func (m *MemStorage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return m.GetAllMetrics(ctx), nil
}
