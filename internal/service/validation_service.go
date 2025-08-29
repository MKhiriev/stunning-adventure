package service

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type ValidatingMetricsService struct {
	inner     MetricsSaverService
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
		return models.Metrics{}, err
	}
	return v.inner.Save(ctx, metric)
}

func (v *ValidatingMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	v.log.Info().Str("func", "*ValidatingMetricsService.SaveAll").Any("metrics", metrics).Msg("validation before SaveAll() started")
	if err := v.validator.Validate(ctx, metrics); err != nil {
		return err
	}
	return v.inner.SaveAll(ctx, metrics)
}

func (v *ValidatingMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, bool) {
	v.log.Info().Str("func", "*ValidatingMetricsService.Get").Any("metrics", metric).Msg("validation before Get() started")
	return v.inner.Get(ctx, metric)
}

func (v *ValidatingMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	v.log.Info().Str("func", "*ValidatingMetricsService.GetAll").Msg("no validation is needed")
	return v.inner.GetAll(ctx)
}

func (v *ValidatingMetricsService) Wrap(wrapper MetricsSaverService) MetricsSaverService {
	v.log.Info().Str("func", "*ValidatingMetricsService.Wrap").Msg("wrapping a service")
	v.inner = wrapper
	return v
}
