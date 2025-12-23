package service

import (
	"context"
	"testing"

	"github.com/MKhiriev/stunning-adventure/models"
)

type mockMetricsService struct {
	saveCalled    bool
	saveAllCalled bool
	getCalled     bool
	getAllCalled  bool

	saveFn    func(context.Context, *models.Metrics) (models.Metrics, error)
	saveAllFn func(context.Context, []models.Metrics) error
	getFn     func(context.Context, *models.Metrics) (models.Metrics, error)
	getAllFn  func(context.Context) ([]models.Metrics, error)
}

func TestValidatingMetricsService_Save_ValidationError(t *testing.T) {
	ctx := context.Background()

	mockInner := &mockMetricsService{}
	svc := NewValidatingMetricsService().Wrap(mockInner)

	metric := &models.Metrics{
		ID: "", // invalid
	}

	_, err := svc.Save(ctx, metric)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if mockInner.saveCalled {
		t.Fatal("inner Save must NOT be called on validation error")
	}
}

func TestValidatingMetricsService_Save_OK(t *testing.T) {
	ctx := context.Background()

	mockInner := &mockMetricsService{
		saveFn: func(ctx context.Context, m *models.Metrics) (models.Metrics, error) {
			return *m, nil
		},
	}

	svc := NewValidatingMetricsService().Wrap(mockInner)

	metric := &models.Metrics{
		ID:    "cpu",
		MType: models.Gauge,
		Value: ptrFloat64(1.23),
	}

	_, err := svc.Save(ctx, metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockInner.saveCalled {
		t.Fatal("inner Save must be called")
	}
}

func TestValidatingMetricsService_SaveAll_ValidationError(t *testing.T) {
	ctx := context.Background()

	mockInner := &mockMetricsService{}
	svc := NewValidatingMetricsService().Wrap(mockInner)

	metrics := []models.Metrics{
		{
			ID:    "ok",
			MType: models.Gauge,
			Value: ptrFloat64(1),
		},
		{
			ID: "", // invalid
		},
	}

	err := svc.SaveAll(ctx, metrics)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if mockInner.saveAllCalled {
		t.Fatal("inner SaveAll must NOT be called")
	}
}

func TestValidatingMetricsService_Get_OK(t *testing.T) {
	ctx := context.Background()

	mockInner := &mockMetricsService{
		getFn: func(ctx context.Context, m *models.Metrics) (models.Metrics, error) {
			return *m, nil
		},
	}

	svc := NewValidatingMetricsService().Wrap(mockInner)

	metric := &models.Metrics{
		ID:    "cpu",
		MType: models.Gauge,
	}

	_, err := svc.Get(ctx, metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockInner.getCalled {
		t.Fatal("inner Get must be called")
	}
}

func TestValidatingMetricsService_GetAll(t *testing.T) {
	ctx := context.Background()

	mockInner := &mockMetricsService{
		getAllFn: func(ctx context.Context) ([]models.Metrics, error) {
			return []models.Metrics{}, nil
		},
	}

	svc := NewValidatingMetricsService().Wrap(mockInner)

	_, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mockInner.getAllCalled {
		t.Fatal("inner GetAll must be called")
	}
}

func (m *mockMetricsService) Save(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	m.saveCalled = true
	return m.saveFn(ctx, metric)
}

func (m *mockMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	m.saveAllCalled = true
	return m.saveAllFn(ctx, metrics)
}

func (m *mockMetricsService) Get(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	m.getCalled = true
	return m.getFn(ctx, metric)
}

func (m *mockMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	m.getAllCalled = true
	return m.getAllFn(ctx)
}

func ptrInt64(v int64) *int64       { return &v }
func ptrFloat64(v float64) *float64 { return &v }
