package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

func TestFileStorage_SaveSingleMetric(t *testing.T) {
	fs := newTestFileStorage(t, false)
	ctx := context.Background()

	metric := models.Metrics{
		ID:    "m1",
		MType: models.Counter,
		Delta: ptrInt64(10),
	}

	// Save single metric
	if _, err := fs.Save(ctx, &metric); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load from file
	loaded, err := fs.GetAll(ctx)
	if err != nil {
		t.Fatalf("LoadMetricsFromFile failed: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(loaded))
	}

	// Get metric
	got, err := fs.Get(ctx, &metric)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != metric.ID || *got.Delta != *metric.Delta {
		t.Fatalf("Get returned unexpected metric: %+v", got)
	}
}

func TestFileStorage_SaveMultipleMetrics(t *testing.T) {
	fs := newTestFileStorage(t, false)
	ctx := context.Background()

	metric1 := models.Metrics{
		ID:    "m1",
		MType: models.Counter,
		Delta: ptrInt64(10),
	}
	metric2 := models.Metrics{
		ID:    "m2",
		MType: models.Gauge,
		Value: ptrFloat64(3.14),
	}

	// Save multiple metrics
	if err := fs.SaveAll(ctx, []models.Metrics{metric1, metric2}); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	// Load from file
	loaded, err := fs.LoadMetricsFromFile(ctx)
	if err != nil {
		t.Fatalf("LoadMetricsFromFile failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(loaded))
	}

	// GetAll
	all, err := fs.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("GetAll returned %d metrics, expected 2", len(all))
	}
}

func TestFileStorage_LoadEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	logger := zerolog.Nop()
	mem := NewMemStorage(&logger)
	cfg := &config.ServerConfig{
		FileStoragePath:        tmpDir,
		RestoreMetricsFromFile: false,
	}
	fs := &FileStorage{
		memStorage:   mem,
		cfg:          cfg,
		log:          &logger,
		fullFileName: filepath.Join(tmpDir, "empty.log"),
	}

	loaded, err := fs.LoadMetricsFromFile(context.Background())
	if err != nil {
		t.Fatalf("LoadMetricsFromFile failed: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected empty slice, got %v", loaded)
	}
}

func TestFileStorage_NewFileStorage_ErrorNoPath(t *testing.T) {
	log := zerolog.Nop()
	mem := NewMemStorage(&log)
	cfg := &config.ServerConfig{FileStoragePath: ""}
	_, err := NewFileStorage(context.Background(), mem, cfg, &log)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestFileStorage_DirectoryError(t *testing.T) {
	log := zerolog.Nop()
	mem := NewMemStorage(&log)
	// point to a file instead of directory to provoke MkdirAll error
	tmpFile := filepath.Join(t.TempDir(), "file.log")
	os.WriteFile(tmpFile, []byte{}, 0644)

	cfg := &config.ServerConfig{FileStoragePath: tmpFile}
	_, err := NewFileStorage(context.Background(), mem, cfg, &log)
	if err == nil {
		t.Fatal("expected error creating directory over file")
	}
}

func newTestFileStorage(t *testing.T, restore bool) *FileStorage {
	t.Helper()

	tmpDir := t.TempDir()
	logger := zerolog.Nop()
	mem := NewMemStorage(&logger)
	cfg := &config.ServerConfig{
		FileStoragePath:        tmpDir,
		RestoreMetricsFromFile: restore,
	}

	fs, err := NewFileStorage(context.Background(), mem, cfg, &logger)
	if err != nil {
		t.Fatalf("failed to create FileStorage: %v", err)
	}
	return fs
}
