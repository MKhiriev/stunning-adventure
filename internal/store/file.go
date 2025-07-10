package store

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"log"
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

	if err := os.MkdirAll(path.Dir(fs.fullFileName), 0755); err != nil {
		log.Println("SaveMetricsToFile: cannot create directory:", err)
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

	err = os.WriteFile(fs.fullFileName, jsonData, 0666)
	log.Println("CREATED FILE err: ", err)
	return err
	//return os.WriteFile(fs.cfg.FileStoragePath+fs.fileName, jsonData, 0666)
}

func (fs *FileStorage) LoadMetricsFromFile() ([]models.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	file, err := os.OpenFile(fs.fullFileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("LoadMetricsFromFile: error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("LoadMetricsFromFile: error stating file: %v", err)
		return nil, err
	}
	if fileInfo.Size() == 0 {
		log.Println("LoadMetricsFromFile: file is empty, returning empty slice")
		return []models.Metrics{}, nil
	}

	data, err := os.ReadFile(fs.fullFileName)
	if err != nil {
		log.Printf("LoadMetricsFromFile: %v", err)
		if os.IsNotExist(err) {
			log.Printf("LoadMetricsFromFile: FILE NOT EXISTS ERROR %v", err)
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		log.Printf("LoadMetricsFromFile: error unmarshalling json: %v", err)
		return nil, err
	}

	return loadedMetrics, nil
}
