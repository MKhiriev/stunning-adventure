package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

func TestCacheMetricsService_Save_CacheOnly_OK(t *testing.T) {
	ctx := context.Background()
	expected := models.Metrics{ID: "m1"}

	cache := &mockStorage{
		saveFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return expected, nil
		},
	}

	svc := newTestCacheService(cache, nil)

	result, err := svc.Save(ctx, &expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cache.saveCalled {
		t.Fatal("cache.Save was not called")
	}

	if result != expected {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCacheMetricsService_Save_CacheError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("cache save failed")

	cache := &mockStorage{
		saveFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return models.Metrics{}, expectedErr
		},
	}

	svc := newTestCacheService(cache, nil)

	_, err := svc.Save(ctx, &models.Metrics{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCacheMetricsService_Save_WithFile_OK(t *testing.T) {
	ctx := context.Background()
	metric := models.Metrics{ID: "m1"}
	all := []models.Metrics{metric}

	cache := &mockStorage{
		saveFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return metric, nil
		},
		getAllFn: func(_ context.Context) ([]models.Metrics, error) {
			return all, nil
		},
	}

	file := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return nil
		},
	}

	svc := newTestCacheService(cache, file)

	_, err := svc.Save(ctx, &metric)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cache.getAllCalled {
		t.Fatal("cache.GetAll was not called")
	}

	if !file.saveAllCalled {
		t.Fatal("file.SaveAll was not called")
	}
}

func TestCacheMetricsService_Save_FileError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("file save failed")

	cache := &mockStorage{
		saveFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return models.Metrics{ID: "m1"}, nil
		},
		getAllFn: func(_ context.Context) ([]models.Metrics, error) {
			return []models.Metrics{{ID: "m1"}}, nil
		},
	}

	file := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return expectedErr
		},
	}

	svc := newTestCacheService(cache, file)

	_, err := svc.Save(ctx, &models.Metrics{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCacheMetricsService_SaveAll_CacheError(t *testing.T) {
	ctx := context.Background()

	cache := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return errors.New("cache error")
		},
	}

	svc := newTestCacheService(cache, nil)

	err := svc.SaveAll(ctx, []models.Metrics{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCacheMetricsService_SaveAll_WithFile_OK(t *testing.T) {
	ctx := context.Background()
	all := []models.Metrics{{ID: "m1"}}

	cache := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return nil
		},
		getAllFn: func(_ context.Context) ([]models.Metrics, error) {
			return all, nil
		},
	}

	file := &mockStorage{
		saveAllFn: func(_ context.Context, _ []models.Metrics) error {
			return nil
		},
	}

	svc := newTestCacheService(cache, file)

	err := svc.SaveAll(ctx, all)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCacheMetricsService_Get(t *testing.T) {
	ctx := context.Background()
	expected := models.Metrics{ID: "m1"}

	cache := &mockStorage{
		getFn: func(_ context.Context, _ *models.Metrics) (models.Metrics, error) {
			return expected, nil
		},
	}

	svc := newTestCacheService(cache, nil)

	result, err := svc.Get(ctx, &models.Metrics{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != expected {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func newTestCacheService(cache, file store.Storage) *CacheMetricsService {
	logger := zerolog.Nop()
	return &CacheMetricsService{
		cache: cache,
		file:  file,
		log:   &logger,
	}
}
