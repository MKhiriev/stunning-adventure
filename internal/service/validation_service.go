package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type ValidatingMetricsService struct {
	inner     MetricsService
	validator validators.Validator
	log       *zerolog.Logger
}

func NewValidatingMetricsService(log *zerolog.Logger) MetricsServiceWrapper {
	return &ValidatingMetricsService{
		validator: validators.NewMetricsValidator(),
		log:       log,
	}
}

func (v *ValidatingMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	v.log.Info().Str("func", "*ValidatingMetricsService.Save").Any("metric", metric).Msg("validation before Save() started")

	if err := v.validator.Validate(ctx, metric); err != nil {
		v.log.Info().Str("func", "*ValidatingMetricsService.Save").Any("metric", metric).Msg("metric is not valid")
		return models.Metrics{}, fmt.Errorf("error during metric validation before saving: %w", err)
	}

	v.log.Info().Str("func", "*ValidatingMetricsService.Save").Any("metric", metric).Msg("metric is valid")

	return v.inner.Save(ctx, metric)
}

func (v *ValidatingMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	v.log.Info().Str("func", "*ValidatingMetricsService.SaveAll").Any("metrics", metrics).Msg("validation before SaveAll() started")

	for _, metric := range metrics {
		if err := v.validator.Validate(ctx, metric); err != nil {
			v.log.Err(err).Str("func", "*ValidatingMetricsService.SaveAll").Any("metrics", metric).Msg("metric is not valid")
			return fmt.Errorf("error during metric validation before saving: %w", err)
		}
	}

	v.log.Info().Str("func", "*ValidatingMetricsService.SaveAll").Any("metrics", metrics).Msg("metric is valid")

	return v.inner.SaveAll(ctx, metrics)
}

func (v *ValidatingMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	v.log.Info().Str("func", "*ValidatingMetricsService.Get").Any("metrics", metric).Msg("validation before Get() started")

	if err := v.validator.Validate(ctx, metric, validators.ID, validators.MType); err != nil {
		v.log.Info().Str("func", "*ValidatingMetricsService.Get").Any("metrics", metric).Msg("metric is not valid")
		return models.Metrics{}, fmt.Errorf("error during metric validation before saving: %w", err)
	}

	v.log.Info().Str("func", "*ValidatingMetricsService.Get").Any("metrics", metric).Msg("metric is valid")

	return v.inner.Get(ctx, metric)
}

func (v *ValidatingMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	v.log.Info().Str("func", "*ValidatingMetricsService.GetAll").Msg("no validation is needed")
	return v.inner.GetAll(ctx)
}

func (v *ValidatingMetricsService) Wrap(wrapper MetricsService) MetricsService {
	v.log.Info().Str("func", "*ValidatingMetricsService.Wrap").Msg("wrapping a service")
	v.inner = wrapper
	return v
}
