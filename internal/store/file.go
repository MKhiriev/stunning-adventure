package store

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"os"
	"path"
)

type MetricsFileStorage interface {
	SaveMetricsToFile([]models.Metrics) error
	LoadMetricsFromFile() ([]models.Metrics, error)
}

type FileStorage struct {
	cfg *config.ServerConfig
	*MemStorage
	fullFileName string
}

func NewFileStorage(memStorage *MemStorage, cfg *config.ServerConfig) *FileStorage {
	fs := &FileStorage{
		MemStorage:   memStorage,
		cfg:          cfg,
		fullFileName: path.Join(cfg.FileStoragePath, "metrics.log"),
	}

	// create directory for metrics file
	_ = os.MkdirAll(path.Dir(fs.fullFileName), 0755)

	// load metrics from file if needed
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

	return os.WriteFile(fs.fullFileName, jsonData, 0644)
}

func (fs *FileStorage) LoadMetricsFromFile() ([]models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// open existing file or create new
	file, err := os.OpenFile(fs.fullFileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// check file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	// if empty - return empty slice of metrics
	if fileInfo.Size() == 0 {
		return []models.Metrics{}, nil
	}

	// if not empty - read file
	data, err := os.ReadFile(fs.fullFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	// decode contents of file from JSON array to slice of metrics
	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		return nil, err
	}

	return loadedMetrics, nil
}
