package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
)

type ValidatingMetricsService struct {
	inner     MetricsService
	validator validators.Validator
}

func NewValidatingMetricsService() MetricsServiceWrapper {
	return &ValidatingMetricsService{
		validator: validators.NewMetricsValidator(),
	}
}

func (v *ValidatingMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if err := v.validator.Validate(ctx, metric); err != nil {
		return models.Metrics{}, fmt.Errorf("error during metric validation before saving: %w", err)
	}

	return v.inner.Save(ctx, metric)
}

func (v *ValidatingMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		if err := v.validator.Validate(ctx, metric); err != nil {
			return fmt.Errorf("error during metric validation before saving: %w", err)
		}
	}

	return v.inner.SaveAll(ctx, metrics)
}

func (v *ValidatingMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if err := v.validator.Validate(ctx, metric, validators.ID, validators.MType); err != nil {
		return models.Metrics{}, fmt.Errorf("error during metric validation before saving: %w", err)
	}

	return v.inner.Get(ctx, metric)
}

func (v *ValidatingMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return v.inner.GetAll(ctx)
}

func (v *ValidatingMetricsService) Wrap(wrapper MetricsService) MetricsService {
	v.inner = wrapper
	return v
}
