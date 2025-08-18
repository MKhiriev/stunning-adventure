package service

import (
	"context"
	"errors"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type DatabaseMetricsService struct {
	db  store.Storage // primary chosen store.Storage provider - db
	log *zerolog.Logger
}

type CacheMetricsService struct {
	file          store.Storage // secondary chosen store.Storage provider - file
	cache         store.Storage // primary chosen store.Storage provider - cache
	storeInterval int64         // for file-storage. Only needed if file storage is provided
	log           *zerolog.Logger
}

func NewMetricsService(fileStorage *store.FileStorage, dbStorage *store.DB, cacheStorage *store.MemStorage, cfg *config.ServerConfig, log *zerolog.Logger) (MetricsSaverService, error) {
	if cfg.DatabaseDSN != "" && dbStorage != nil {
		return NewDatabaseMetricsService(dbStorage, log, cfg), nil
	}

	if cfg.FileStoragePath != "" && fileStorage != nil && cacheStorage != nil {
		return NewCacheMetricsService(cacheStorage, fileStorage, log, cfg), nil
	}
	if cacheStorage != nil {
		return NewCacheMetricsService(cacheStorage, nil, log, cfg), nil
	}

	log.Error().Str("func", "service.NewMetricsService").Msg("error creating DatabaseMetricsService: nil db was provided")
	return nil, errors.New("error creating DatabaseMetricsService: nil db was provided")
}

func NewDatabaseMetricsService(dbStorage *store.DB, log *zerolog.Logger, cfg *config.ServerConfig) *DatabaseMetricsService {
	log.Info().Str("func", "service.NewDatabaseMetricsService").Msg("metrics service with DB was created")
	return &DatabaseMetricsService{
		log: log,
		db:  dbStorage,
	}
}

func NewCacheMetricsService(cache *store.MemStorage, fileStorage *store.FileStorage, log *zerolog.Logger, cfg *config.ServerConfig) *CacheMetricsService {
	if fileStorage != nil {
		log.Info().Str("func", "service.NewCacheMetricsService").Msg("metrics service with File and Cache was created")
		// TODO add start of goroutine for saving metrics to file in later sprints
		return &CacheMetricsService{
			cache:         cache,
			file:          fileStorage,
			log:           log,
			storeInterval: cfg.StoreInterval,
		}
	}

	log.Info().Str("func", "service.NewCacheMetricsService").Msg("metrics service with only Cache was created")
	return &CacheMetricsService{
		cache: cache,
		log:   log,
	}
}

func (m *DatabaseMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return m.db.Save(ctx, metric)
}

func (m *DatabaseMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	return m.db.SaveAll(ctx, metrics)
}

// Get Возвращает метрику по имени и типу.// Получает данные из главного хранилища - Память или БД
func (m *DatabaseMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, bool) {
	return m.db.Get(ctx, metric)
}

// GetAll Возвращает все записанные в главное хранилище метрики.// Получает данные из главного хранилища - Память или БД
func (m *DatabaseMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return m.db.GetAll(ctx)
}

func (c *CacheMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	// save metric in cache memory
	result, err := c.cache.Save(ctx, metric)
	if err != nil {
		return models.Metrics{}, err
	}

	// if file storage was provided
	if c.file != nil {
		// get all metrics from Cache storage
		allMetrics, err := c.cache.GetAll(ctx)
		if err != nil {
			return models.Metrics{}, err
		}

		// save all metrics to file
		err = c.file.SaveAll(ctx, allMetrics)
		if err != nil {
			return models.Metrics{}, err
		}
	}

	return result, nil
}

func (c *CacheMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	// save all metrics in cache memory
	err := c.cache.SaveAll(ctx, metrics)
	if err != nil {
		return err
	}

	// if file storage was provided
	if c.file != nil {
		// save all provided metrics
		err = c.file.SaveAll(ctx, metrics)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CacheMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, bool) {
	return c.cache.Get(ctx, metric)
}

func (c *CacheMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return c.cache.GetAll(ctx)
}
