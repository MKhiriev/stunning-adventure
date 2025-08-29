package validators

import (
	"context"
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
			return ErrUnsupportedType
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
				return ErrEmptyID
			}
		case "type":
			if metric.MType == "" {
				return ErrEmptyType
			}
			if !slices.Contains(v.allowedMetricTypes, metric.MType) {
				return ErrInvalidType
			}
		case "value":
			if metric.Value == nil && metric.Delta == nil {
				return ErrNoValue
			}
		default:
			return ErrUnknownField
		}
	}
	return nil
}
