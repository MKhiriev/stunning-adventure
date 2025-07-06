package services

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"time"
)

type MetricsService interface {
	AddCounter(models.Metrics) (models.Metrics, error)
	UpdateGauge(models.Metrics) (models.Metrics, error)
	GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool)
	GetAllMetrics() []models.Metrics
	SaveMetricsToFile() error
	LoadMetricsFromFile() ([]models.Metrics, error)
}

type ServerMetricsService struct {
	MemStorage    *store.MemStorage
	FileStorage   *store.FileStorage
	StoreInterval int64
	Restore       bool
}

func NewMetricsService(memStorage *store.MemStorage, fileStorage *store.FileStorage, storeInterval int64, loadMetricsFromFile bool) *ServerMetricsService {
	ms := &ServerMetricsService{
		MemStorage:    memStorage,
		FileStorage:   fileStorage,
		StoreInterval: storeInterval,
		Restore:       loadMetricsFromFile,
	}

	if loadMetricsFromFile {
		metricsFromFile, err := ms.LoadMetricsFromFile()
		if err != nil {
			return nil
		}
		for _, metric := range metricsFromFile {
			ms.MemStorage.Memory[metric.ID] = metric
		}
	}

	return ms
}

func (s *ServerMetricsService) AddCounter(metrics models.Metrics) (models.Metrics, error) {
	return s.MemStorage.AddCounter(metrics)
}

func (s *ServerMetricsService) UpdateGauge(metrics models.Metrics) (models.Metrics, error) {
	return s.MemStorage.UpdateGauge(metrics)
}

func (s *ServerMetricsService) GetMetricByNameAndType(metricName string, metricType string) (models.Metrics, bool) {
	return s.MemStorage.GetMetricByNameAndType(metricName, metricType)
}

func (s *ServerMetricsService) GetAllMetrics() []models.Metrics {
	return s.MemStorage.GetAllMetrics()
}

func (s *ServerMetricsService) SaveMetricsToFile(metrics []models.Metrics) error {
	return s.FileStorage.SaveMetrics(metrics)
}

func (s *ServerMetricsService) LoadMetricsFromFile() ([]models.Metrics, error) {
	return s.FileStorage.LoadMetrics()
}

func (s *ServerMetricsService) SaveMetricsToFilePeriodically() {
	utils.RunWithTicker(func() {
		metrics := s.GetAllMetrics()
		if metrics != nil {
			_ = s.FileStorage.SaveMetrics(metrics)
		}
	}, time.Duration(s.StoreInterval)*time.Second)
}
