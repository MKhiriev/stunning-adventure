package validators

import (
	"context"
	"errors"
	"slices"

	"github.com/MKhiriev/stunning-adventure/models"
)

const (
	ID    = "id"
	MType = "type"
	Value = "value"
	Delta = "value"
)

type MetricsValidator struct {
	allowedMetricTypes []string
}

func NewMetricsValidator() *MetricsValidator {
	return &MetricsValidator{
		allowedMetricTypes: []string{models.Gauge, models.Counter},
	}
}

func (v *MetricsValidator) Validate(ctx context.Context, obj any, fields ...string) error {
	metric, ok := obj.(models.Metrics)
	if !ok {
		// check if it's a pointer
		ptr, ok := obj.(*models.Metrics)
		if !ok {
			return errors.New("unsupported type for validation")
		}
		metric = *ptr
	}

	// if empty fields list - validate all!
	if len(fields) == 0 {
		fields = []string{"id", "type", "value"}
	}

	for _, f := range fields {
		switch f {
		case "id":
			if metric.ID == "" {
				return errors.New("metric id is empty")
			}
		case "type":
			if metric.MType == "" {
				return errors.New("metric type is empty")
			}
			if !slices.Contains(v.allowedMetricTypes, metric.MType) {
				return errors.New("metric type is not valid")
			}
		case "value":
			if metric.Value == nil && metric.Delta == nil {
				return errors.New("metric has no value")
			}
		default:
			return errors.New("unknown field for validation: " + f)
		}
	}
	return nil
}
