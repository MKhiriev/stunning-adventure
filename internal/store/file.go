package store

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"os"
)

type MetricsFileStorage interface {
	SaveMetrics([]models.Metrics) error
	LoadMetrics() ([]models.Metrics, error)
}

type FileStorage struct {
	cfg *config.ServerConfig
	*MemStorage
}

func NewFileStorage(memStorage *MemStorage, cfg *config.ServerConfig) *FileStorage {
	return &FileStorage{
		MemStorage: memStorage,
		cfg:        cfg,
	}
}

func (fs *FileStorage) SaveMetrics(allMetrics []models.Metrics) error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	jsonData, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}

	return os.WriteFile(fs.cfg.FileStoragePath, jsonData, 0644)
}

func (fs *FileStorage) LoadMetrics() ([]models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := os.ReadFile(fs.cfg.FileStoragePath)
	if err != nil {
		return nil, err
	}

	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		return nil, err
	}

	return loadedMetrics, nil
}
