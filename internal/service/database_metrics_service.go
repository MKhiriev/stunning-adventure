package service

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type DatabaseMetricsService struct {
	db  store.Storage // primary chosen store.Storage provider - db
	log *zerolog.Logger
}

func (m *DatabaseMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return m.db.Save(ctx, metric)
}

func (m *DatabaseMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	return m.db.SaveAll(ctx, metrics)
}

// Get Возвращает метрику по имени и типу.// Получает данные из главного хранилища - Память или БД
func (m *DatabaseMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return m.db.Get(ctx, metric)
}

// GetAll Возвращает все записанные в главное хранилище метрики.// Получает данные из главного хранилища - Память или БД
func (m *DatabaseMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return m.db.GetAll(ctx)
}
