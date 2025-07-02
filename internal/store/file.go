package store

import (
	"encoding/json"
	"github.com/MKhiriev/stunning-adventure/models"
	"os"
	"path"
)

type MetricsFileStorage interface {
	SaveMetrics([]models.Metrics) error
	LoadMetrics() ([]models.Metrics, error)
}

type FileStorage struct {
	File     *os.File
	filePath string
}

func NewFileStorage(fileName, fileStoragePath string) (*FileStorage, error) {
	filePath := path.Join(fileStoragePath, fileName)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		File:     file,
		filePath: filePath,
	}, nil
}

func (fs *FileStorage) SaveMetrics(allMetrics []models.Metrics) error {
	jsonData, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}

	_, err = fs.File.Write(jsonData)
	return err
}

func (fs *FileStorage) LoadMetrics() ([]models.Metrics, error) {
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		return nil, err
	}

	return loadedMetrics, nil
}
