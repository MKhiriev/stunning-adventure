package validators

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/models"
)

func TestMetricsValidator_Validate(t *testing.T) {
	ctx := context.Background()
	v := NewMetricsValidator()

	validGaugeValue := float64(10)
	validCounterDelta := int64(5)

	tests := []struct {
		name    string
		obj     any
		fields  []string
		wantErr error
	}{
		// --- type checks ---
		{
			name:    "unsupported type",
			obj:     "not a metric",
			wantErr: ErrUnsupportedType,
		},
		{
			name: "pointer metric supported",
			obj: &models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
				Value: &validGaugeValue,
			},
			wantErr: nil,
		},

		// --- empty fields list (validate all) ---
		{
			name: "valid gauge metric",
			obj: models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
				Value: &validGaugeValue,
			},
			wantErr: nil,
		},
		{
			name: "valid counter metric",
			obj: models.Metrics{
				ID:    "requests",
				MType: models.Counter,
				Delta: &validCounterDelta,
			},
			wantErr: nil,
		},

		// --- id validation ---
		{
			name: "empty id",
			obj: models.Metrics{
				ID:    "",
				MType: models.Gauge,
				Value: &validGaugeValue,
			},
			wantErr: ErrEmptyID,
		},
		{
			name: "validate only id",
			obj: models.Metrics{
				ID: "only-id",
			},
			fields:  []string{"id"},
			wantErr: nil,
		},

		// --- type validation ---
		{
			name: "empty type",
			obj: models.Metrics{
				ID: "cpu",
			},
			fields:  []string{"type"},
			wantErr: ErrEmptyType,
		},
		{
			name: "invalid type",
			obj: models.Metrics{
				ID:    "cpu",
				MType: "histogram",
			},
			fields:  []string{"type"},
			wantErr: ErrInvalidType,
		},

		// --- value validation ---
		{
			name: "no value and no delta",
			obj: models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
			},
			fields:  []string{"value"},
			wantErr: ErrNoValue,
		},
		{
			name: "value present",
			obj: models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
				Value: &validGaugeValue,
			},
			fields:  []string{"value"},
			wantErr: nil,
		},
		{
			name: "delta present",
			obj: models.Metrics{
				ID:    "req",
				MType: models.Counter,
				Delta: &validCounterDelta,
			},
			fields:  []string{"value"},
			wantErr: nil,
		},

		// --- mixed fields ---
		{
			name: "partial validation success",
			obj: models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
			},
			fields:  []string{"id", "type"},
			wantErr: nil,
		},

		// --- unknown field ---
		{
			name: "unknown field",
			obj: models.Metrics{
				ID:    "cpu",
				MType: models.Gauge,
			},
			fields:  []string{"unknown"},
			wantErr: ErrUnknownField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(ctx, tt.obj, tt.fields...)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
