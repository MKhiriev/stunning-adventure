package service

import (
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog"
)

type testWrapper struct {
	wrapped bool
}

func TestMetricsServiceBuilder_DB_Priority(t *testing.T) {
	cfg := &config.ServerConfig{
		DatabaseDSN: "postgres://test",
	}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithDB(&store.DB{}).
		WithCache(&store.MemStorage{}).
		WithFile(&store.FileStorage{})

	svc, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := svc.(*DatabaseMetricsService); !ok {
		t.Fatalf("expected DatabaseMetricsService, got %T", svc)
	}
}

func TestMetricsServiceBuilder_DB_NilStorage_Error(t *testing.T) {
	cfg := &config.ServerConfig{
		DatabaseDSN: "postgres://test",
	}

	builder := NewMetricsServiceBuilder(cfg, testLogger())

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricsServiceBuilder_FileAndCache(t *testing.T) {
	cfg := &config.ServerConfig{
		FileStoragePath: "/tmp",
		StoreInterval:   10,
	}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithFile(&store.FileStorage{}).
		WithCache(&store.MemStorage{})

	svc, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := svc.(*CacheMetricsService); !ok {
		t.Fatalf("expected CacheMetricsService, got %T", svc)
	}
}

func TestMetricsServiceBuilder_FileWithoutCache_Error(t *testing.T) {
	cfg := &config.ServerConfig{
		FileStoragePath: "/tmp",
	}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithFile(&store.FileStorage{})

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricsServiceBuilder_CacheOnly(t *testing.T) {
	cfg := &config.ServerConfig{}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithCache(&store.MemStorage{})

	svc, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := svc.(*CacheMetricsService); !ok {
		t.Fatalf("expected CacheMetricsService, got %T", svc)
	}
}

func TestMetricsServiceBuilder_NoStorage_Error(t *testing.T) {
	cfg := &config.ServerConfig{}

	builder := NewMetricsServiceBuilder(cfg, testLogger())

	_, err := builder.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMetricsServiceBuilder_WrapperApplied(t *testing.T) {
	cfg := &config.ServerConfig{}

	wrapper := &testWrapper{}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithCache(&store.MemStorage{}).
		WithWrapper(wrapper)

	_, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !wrapper.wrapped {
		t.Fatal("expected wrapper to be applied")
	}
}

func TestMetricsServiceBuilder_MultipleWrappers(t *testing.T) {
	cfg := &config.ServerConfig{}

	w1 := &testWrapper{}
	w2 := &testWrapper{}

	builder := NewMetricsServiceBuilder(cfg, testLogger()).
		WithCache(&store.MemStorage{}).
		WithWrapper(w1).
		WithWrapper(w2)

	_, err := builder.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !w1.wrapped || !w2.wrapped {
		t.Fatal("expected all wrappers to be applied")
	}
}

func (w *testWrapper) Wrap(inner MetricsService) MetricsService {
	w.wrapped = true
	return inner
}

func testLogger() *zerolog.Logger {
	l := zerolog.Nop()
	return &l
}
