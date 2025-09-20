package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type FileStorage struct {
	cfg          *config.ServerConfig
	memStorage   *MemStorage
	log          *zerolog.Logger
	fullFileName string
}

func NewFileStorage(memStorage *MemStorage, cfg *config.ServerConfig, log *zerolog.Logger) (*FileStorage, error) {
	if cfg.FileStoragePath == "" {
		log.Error().Msg("no file storage path was provided")
		return nil, errors.New("no file storage path was provided")
	}

	fs := &FileStorage{
		memStorage:   memStorage,
		cfg:          cfg,
		log:          log,
		fullFileName: path.Join(cfg.FileStoragePath, "metrics.log"),
	}

	// create directory for metrics file
	if err := os.MkdirAll(path.Dir(fs.fullFileName), 0755); err != nil {
		fs.log.Err(err).Str("func", "store.NewFileStorage").Msg("error creating directory for metrics file")
		return nil, err
	}

	// load metrics from file if needed
	if cfg.RestoreMetricsFromFile {
		metricsFromFile, err := fs.LoadMetricsFromFile(context.TODO())
		if err != nil {
			fs.log.Err(err).Str("func", "store.NewFileStorage").Msg("error loading metrics from file")
			return nil, err
		}
		for _, metric := range metricsFromFile {
			fs.memStorage.Memory[metric.ID] = metric
		}
	}

	return fs, nil
}

func (fs *FileStorage) SaveMetricsToFile(ctx context.Context, allMetrics []models.Metrics) error {
	fs.memStorage.mu.Lock()
	defer fs.memStorage.mu.Unlock()

	jsonData, err := json.Marshal(allMetrics)
	if err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.SaveMetricsToFile").Msg("error marshalling metric to JSON")
		return err
	}

	return os.WriteFile(fs.fullFileName, jsonData, 0644)
}

func (fs *FileStorage) LoadMetricsFromFile(context.Context) ([]models.Metrics, error) {
	fs.memStorage.mu.Lock()
	defer fs.memStorage.mu.Unlock()

	// open existing file or create new
	file, err := os.OpenFile(fs.fullFileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.LoadMetricsFromFile").Msg("error opening existing file or creating new file")
		return nil, err
	}
	defer file.Close()

	// check file size
	fileInfo, err := file.Stat()
	if err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.LoadMetricsFromFile").Msg("error getting file size")
		return nil, err
	}
	// if empty - return empty slice of metrics
	if fileInfo.Size() == 0 {
		fs.log.Info().Str("func", "*FileStorage.LoadMetricsFromFile").Msg("metrics file is empty")
		return []models.Metrics{}, nil
	}

	// if not empty - read file
	data, err := os.ReadFile(fs.fullFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fs.log.Err(err).Str("func", "*FileStorage.LoadMetricsFromFile").Msg("error during reading file")
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	// decode contents of file from JSON array to slice of metrics
	var loadedMetrics []models.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.LoadMetricsFromFile").Msg("error during unmarshalling JSON from file to a slice of metrics")
		return nil, err
	}

	return loadedMetrics, nil
}

func (fs *FileStorage) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	fs.memStorage.mu.Lock()
	defer fs.memStorage.mu.Unlock()

	// save metric to file
	fs.memStorage.Memory[metric.ID] = metric
	if err := fs.SaveMetricsToFile(ctx, fs.memStorage.GetAllMetrics(ctx)); err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.Save").Msg("error during saving metric to a file")
		return models.Metrics{}, err
	}

	return metric, nil
}

func (fs *FileStorage) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	return fs.SaveMetricsToFile(ctx, metrics)
}

func (fs *FileStorage) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	metrics, err := fs.LoadMetricsFromFile(ctx)
	if err != nil {
		fs.log.Err(err).Str("func", "*FileStorage.Get").Msg("error during getting metric from file")
		return models.Metrics{}, err
	}

	for _, m := range metrics {
		if m.ID == metric.ID && m.MType == metric.MType {
			return m, nil
		}
	}

	fs.log.Info().Str("func", "*FileStorage.Get").Msg("no metric was found")
	return models.Metrics{}, ErrNotFound
}

func (fs *FileStorage) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return fs.LoadMetricsFromFile(ctx)
}
