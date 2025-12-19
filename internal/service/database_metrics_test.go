package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type mockStorage struct {
	saveCalled    bool
	saveAllCalled bool
	getCalled     bool
	getAllCalled  bool

	saveFn    func(context.Context, *models.Metrics) (models.Metrics, error)
	saveAllFn func(context.Context, []models.Metrics) error
	getFn     func(context.Context, *models.Metrics) (models.Metrics, error)
	getAllFn  func(context.Context) ([]models.Metrics, error)
}

func TestDatabaseMetricsService_Save_OK(t *testing.T) {
	ctx := context.Background()
	expected := models.Metrics{ID: "m1", MType: "counter"}

	mock := &mockStorage{
		saveFn: func(_ context.Context, m *models.Metrics) (models.Metrics, error) {
			return expected, nil
		},
	}

	svc := newTestDatabaseMetricsService(mock)

	result, err := svc.Save(ctx, &expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.saveCalled {
		t.Fatal("Save was not delegated to storage")
	}

	if result != expected {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDatabaseMetricsService_Save_Error(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("save failed")

	mock := &mockStorage{
		saveFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return models.Metrics{}, expectedErr
		},
	}

	svc := newTestDatabaseMetricsService(mock)

	_, err := svc.Save(ctx, &models.Metrics{})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDatabaseMetricsService_SaveAll_OK(t *testing.T) {
	ctx := context.Background()

	mock := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return nil
		},
	}

	svc := newTestDatabaseMetricsService(mock)

	err := svc.SaveAll(ctx, []models.Metrics{{ID: "m1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.saveAllCalled {
		t.Fatal("SaveAll was not delegated")
	}
}

func TestDatabaseMetricsService_Get_OK(t *testing.T) {
	ctx := context.Background()
	expected := models.Metrics{ID: "m1", MType: "gauge"}

	mock := &mockStorage{
		getFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return expected, nil
		},
	}

	svc := newTestDatabaseMetricsService(mock)

	result, err := svc.Get(ctx, &models.Metrics{ID: "m1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.getCalled {
		t.Fatal("Get was not delegated")
	}

	if result != expected {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDatabaseMetricsService_GetAll_OK(t *testing.T) {
	ctx := context.Background()
	expected := []models.Metrics{{ID: "m1"}, {ID: "m2"}}

	mock := &mockStorage{
		getAllFn: func(_ context.Context) ([]models.Metrics, error) {
			return expected, nil
		},
	}

	svc := newTestDatabaseMetricsService(mock)

	result, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.getAllCalled {
		t.Fatal("GetAll was not delegated")
	}

	if len(result) != len(expected) {
		t.Fatalf("unexpected result length: %d", len(result))
	}
}

func (m *mockStorage) Save(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	m.saveCalled = true
	return m.saveFn(ctx, metric)
}

func (m *mockStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	m.saveAllCalled = true
	return m.saveAllFn(ctx, metrics)
}

func (m *mockStorage) Get(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	m.getCalled = true
	return m.getFn(ctx, metric)
}

func (m *mockStorage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	m.getAllCalled = true
	return m.getAllFn(ctx)
}

func newTestDatabaseMetricsService(storage store.Storage) *DatabaseMetricsService {
	logger := zerolog.Nop()
	return &DatabaseMetricsService{
		db:  storage,
		log: &logger,
	}
}
