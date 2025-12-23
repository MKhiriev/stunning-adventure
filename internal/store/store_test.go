package store

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

func newTestMemStorage() *MemStorage {
	logger := zerolog.Nop() // no-op logger for testing
	return NewMemStorage(&logger)
}

func TestMemStorage_AddCounter(t *testing.T) {
	ctx := context.Background()
	store := newTestMemStorage()

	delta := int64(5)
	metric := &models.Metrics{ID: "counter1", MType: models.Counter, Delta: &delta}

	// first insert
	res, err := store.AddCounter(ctx, metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *res.Delta != delta {
		t.Fatalf("expected delta %d, got %d", delta, *res.Delta)
	}

	// second insert (should sum deltas)
	newDelta := int64(7)
	metric2 := &models.Metrics{ID: "counter1", MType: models.Counter, Delta: &newDelta}

	res2, err := store.AddCounter(ctx, metric2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := delta + newDelta
	if *res2.Delta != expected {
		t.Fatalf("expected delta %d, got %d", expected, *res2.Delta)
	}

	// wrong type
	_, err = store.AddCounter(ctx, &models.Metrics{ID: "x", MType: models.Gauge})
	if err == nil || err.Error() != "metric type is not `counter`" {
		t.Fatalf("expected type error, got %v", err)
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	ctx := context.Background()
	store := newTestMemStorage()

	val := float64(3.14)
	metric := &models.Metrics{ID: "g1", MType: models.Gauge, Value: &val}

	// first insert
	res, err := store.UpdateGauge(ctx, metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *res.Value != val {
		t.Fatalf("expected value %v, got %v", val, *res.Value)
	}

	// second update
	newVal := float64(6.28)
	metric2 := &models.Metrics{ID: "g1", MType: models.Gauge, Value: &newVal}

	res2, err := store.UpdateGauge(ctx, metric2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *res2.Value != newVal {
		t.Fatalf("expected value %v, got %v", newVal, *res2.Value)
	}

	// wrong type
	_, err = store.UpdateGauge(ctx, &models.Metrics{ID: "x", MType: models.Counter})
	if err == nil || err.Error() != "metric type is not `gauge`" {
		t.Fatalf("expected type error, got %v", err)
	}
}

func TestMemStorage_GetMetricByNameAndType(t *testing.T) {
	ctx := context.Background()
	store := newTestMemStorage()

	delta := int64(42)
	counter := &models.Metrics{ID: "c1", MType: models.Counter, Delta: &delta}
	store.AddCounter(ctx, counter)

	val := 3.14
	gauge := &models.Metrics{ID: "g1", MType: models.Gauge, Value: &val}
	store.UpdateGauge(ctx, gauge)

	tests := []struct {
		name      string
		id        string
		mType     string
		wantError bool
	}{
		{"existing_counter", "c1", models.Counter, false},
		{"existing_gauge", "g1", models.Gauge, false},
		{"not_found", "x", models.Counter, true},
		{"wrong_type", "c1", models.Gauge, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := store.GetMetricByNameAndType(ctx, tt.id, tt.mType)
			if tt.wantError {
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("expected ErrNotFound, got %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if m.ID != tt.id || m.MType != tt.mType {
					t.Fatalf("unexpected metric: %+v", m)
				}
			}
		})
	}
}

func TestMemStorage_SaveAndSaveAll(t *testing.T) {
	ctx := context.Background()
	store := newTestMemStorage()

	counter := models.Metrics{ID: "c1", MType: models.Counter, Delta: ptrInt64(5)}
	gauge := models.Metrics{ID: "g1", MType: models.Gauge, Value: ptrFloat64(1.23)}

	// Save
	_, err := store.Save(ctx, &counter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = store.Save(ctx, &gauge)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SaveAll
	metrics := []models.Metrics{
		{ID: "c2", MType: models.Counter, Delta: ptrInt64(7)},
		{ID: "g2", MType: models.Gauge, Value: ptrFloat64(3.21)},
	}
	if err := store.SaveAll(ctx, metrics); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// unsupported type
	bad := models.Metrics{ID: "x", MType: "unknown"}
	if _, err := store.Save(ctx, &bad); err == nil {
		t.Fatal("expected unsupported metric type error")
	}
	if err := store.SaveAll(ctx, []models.Metrics{bad}); err == nil {
		t.Fatal("expected unsupported metric type error")
	}
}

func TestMemStorage_GetAll(t *testing.T) {
	ctx := context.Background()
	store := newTestMemStorage()

	counter := models.Metrics{ID: "c1", MType: models.Counter, Delta: ptrInt64(5)}
	gauge := models.Metrics{ID: "g1", MType: models.Gauge, Value: ptrFloat64(1.23)}
	store.Save(ctx, &counter)
	store.Save(ctx, &gauge)

	all, err := store.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(all))
	}
}

func ptrInt64(v int64) *int64       { return &v }
func ptrFloat64(v float64) *float64 { return &v }
