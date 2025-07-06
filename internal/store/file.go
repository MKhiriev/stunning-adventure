package store

import (
	"encoding/json"
	"errors"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"os"
)

type MetricsFileStorage interface {
	SaveMetricsToFile([]models.Metrics) error
	LoadMetricsFromFile() ([]models.Metrics, error)
}

type FileStorage struct {
	cfg *config.ServerConfig
	*MemStorage
}

func NewFileStorage(memStorage *MemStorage, cfg *config.ServerConfig) *FileStorage {
	fs := &FileStorage{
		MemStorage: memStorage,
		cfg:        cfg,
	}

	if cfg.RestoreMetricsFromFile {
		metricsFromFile, _ := fs.LoadMetricsFromFile()
		for _, metric := range metricsFromFile {
			fs.MemStorage.Memory[metric.ID] = metric
		}
	}

	return fs
}

func (fs *FileStorage) SaveMetricsToFile(allMetrics []models.Metrics) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	jsonData, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.cfg.FileStoragePath, jsonData, 0644)
}

func (fs *FileStorage) LoadMetricsFromFile() ([]models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.cfg.FileStoragePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		return nil, err
	}

	return loadedMetrics, nil
}
