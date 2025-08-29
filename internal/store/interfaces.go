package store

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

type Storage interface {
	Save(context.Context, models.Metrics) (models.Metrics, error) // for MetricsFileStorage we need to save multiple values - modify to accept multiple values
	SaveAll(context.Context, []models.Metrics) error
	Get(context.Context, models.Metrics) (models.Metrics, error)
	GetAll(context.Context) ([]models.Metrics, error)
}

type MetricsFileStorage interface {
	SaveMetricsToFile(context.Context, []models.Metrics) error
	LoadMetricsFromFile(context.Context) ([]models.Metrics, error)
}

type MetricsCacheStorage interface {
	AddCounter(context.Context, models.Metrics) (models.Metrics, error)
	UpdateGauge(context.Context, models.Metrics) (models.Metrics, error)
	GetMetricByNameAndType(ctx context.Context, metricName string, metricType string) (models.Metrics, bool)
	GetAllMetrics(context.Context) []models.Metrics
}

type MetricsDatabaseStorage interface {
	Migrate(context.Context) error
}

// ErrorClassification тип для классификации ошибок
type ErrorClassification int

type ErrorClassificator interface {
	Classify(err error) ErrorClassification
}
