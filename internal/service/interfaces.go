package service

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/internal/service/observer"
	"github.com/MKhiriev/stunning-adventure/models"
)

type MetricsService interface {
	Save(context.Context, models.Metrics) (models.Metrics, error)
	SaveAll(context.Context, []models.Metrics) error
	Get(context.Context, models.Metrics) (models.Metrics, error)
	GetAll(context.Context) ([]models.Metrics, error)
}

type PingService interface {
	Ping(ctx context.Context) error
}

type MetricsServiceWrapper interface {
	Wrap(MetricsService) MetricsService
}

type AuditPublisher interface {
	Register(observer observer.AuditObserver) error
	Deregister(observer observer.AuditObserver) error
	NotifyAll(ctx context.Context, event models.AuditEvent) error
}
