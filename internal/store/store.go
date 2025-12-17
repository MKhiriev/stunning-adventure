package store

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type MemStorage struct {
	Memory map[MetricID]models.Metrics `json:"metrics"`
	mu     *sync.Mutex
	log    *zerolog.Logger
}

type MetricID struct {
	ID    string
	MType string
}

func NewMemStorage(log *zerolog.Logger) *MemStorage {
	return &MemStorage{Memory: make(map[MetricID]models.Metrics), mu: &sync.Mutex{}, log: log}
}

func (m *MemStorage) AddCounter(ctx context.Context, metrics *models.Metrics) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result models.Metrics

	if metrics.MType != models.Counter {
		return models.Metrics{}, errors.New("metric type is not `counter`")
	}

	metricId := MetricID{ID: metrics.ID, MType: models.Counter}
	val, ok := m.Memory[metricId]
	// if metric name exists in storage - apply Counter logic
	if ok {
		newDelta := *val.Delta + *metrics.Delta
		val.Delta = &newDelta

		m.Memory[metricId] = val
		result = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metricId] = *metrics
		result = *metrics
	}

	return result, nil
}

func (m *MemStorage) UpdateGauge(ctx context.Context, metrics *models.Metrics) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result models.Metrics

	if metrics.MType != models.Gauge {
		return models.Metrics{}, errors.New("metric type is not `gauge`")
	}

	metricId := MetricID{ID: metrics.ID, MType: models.Gauge}
	val, ok := m.Memory[metricId]
	// if metric name exists in storage - apply Gauge logic
	if ok {
		val.Value = metrics.Value
		m.Memory[metricId] = val
		result = val
	} else {
		// if metric name doesn't exist - add it
		m.Memory[metricId] = *metrics
		result = *metrics
	}

	return result, nil
}

func (m *MemStorage) GetMetricByNameAndType(ctx context.Context, metricName string, metricType string) (models.Metrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metricId := MetricID{ID: metricName, MType: metricType}
	foundMetric, ok := m.Memory[metricId]
	if ok {
		if foundMetric.MType == metricType && hasValue(foundMetric) {
			foundMetric.MType = metricType
			return foundMetric, nil
		}
		return models.Metrics{}, ErrNotFound
	}

	return models.Metrics{}, ErrNotFound
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slices.Collect(maps.Values(m.Memory))
}

func (m *MemStorage) Save(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	switch metric.MType {
	case models.Counter:
		return m.AddCounter(ctx, metric)
	case models.Gauge:
		return m.UpdateGauge(ctx, metric)
	default:
		return *metric, errors.New("unsupported metric type")
	}
}

func (m *MemStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	var err error
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			_, err = m.AddCounter(ctx, &metric)
			if err != nil {
				return err
			}
		case models.Gauge:
			_, err = m.UpdateGauge(ctx, &metric)
			if err != nil {
				return err
			}
		default:
			return errors.New("unsupported metric type")
		}
	}
	return nil
}

func (m *MemStorage) Get(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	return m.GetMetricByNameAndType(ctx, metric.ID, metric.MType)
}

func (m *MemStorage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return m.GetAllMetrics(ctx), nil
}

func hasValue(metric models.Metrics) bool {
	switch metric.MType {
	case models.Gauge:
		if metric.Value != nil {
			return true
		}
	case models.Counter:
		if metric.Delta != nil {
			return true
		}
	}

	return false
}
