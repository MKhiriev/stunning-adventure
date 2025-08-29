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

type MetricsServiceBuilder struct {
	wrappers     []MetricsServiceWrapper
	fileStorage  *store.FileStorage
	dbStorage    *store.DB
	cacheStorage *store.MemStorage
	cfg          *config.ServerConfig
	log          *zerolog.Logger
}

func NewMetricsServiceBuilder(cfg *config.ServerConfig, log *zerolog.Logger) *MetricsServiceBuilder {
	return &MetricsServiceBuilder{
		cfg: cfg,
		log: log,
	}
}

func (b *MetricsServiceBuilder) WithDB(db *store.DB) *MetricsServiceBuilder {
	b.dbStorage = db
	return b
}

func (b *MetricsServiceBuilder) WithFile(file *store.FileStorage) *MetricsServiceBuilder {
	b.fileStorage = file
	return b
}

func (b *MetricsServiceBuilder) WithCache(cache *store.MemStorage) *MetricsServiceBuilder {
	b.cacheStorage = cache
	return b
}

// WithWrapper add wrappers
func (b *MetricsServiceBuilder) WithWrapper(wrapperService MetricsServiceWrapper) *MetricsServiceBuilder {
	b.wrappers = append(b.wrappers, wrapperService)
	return b
}

func (b *MetricsServiceBuilder) Build() (MetricsSaverService, error) {
	var service MetricsSaverService

	b.log.Info().Str("func", "MetricsServiceBuilder.Build").Msg("started building MetricsSaverService")
	service, err := b.buildMetricsService()
	if err != nil {
		return nil, err
	}

	// adding wrappers if they exist
	if len(b.wrappers) > 0 {
		b.log.Info().Str("func", "MetricsServiceBuilder.Build").Msg("started adding wrapper services for the MetricsSaverService")
		for _, wrapper := range b.wrappers {
			service = wrapper.Wrap(service)
		}
	}

	return service, nil
}

func (b *MetricsServiceBuilder) buildMetricsService() (MetricsSaverService, error) {
	// DB - high priority
	if b.cfg.DatabaseDSN != "" {
		if b.dbStorage == nil {
			b.log.Error().Msg("DB storage is nil")
			return nil, errors.New("db storage is nil")
		}
		b.log.Info().Str("func", "MetricsServiceBuilder.buildMetricsService").Msg("DatabaseMetricsService created")
		return &DatabaseMetricsService{
			db:  b.dbStorage,
			log: b.log,
		}, nil
	}

	// File + Cache
	if b.cfg.FileStoragePath != "" {
		if b.fileStorage == nil {
			b.log.Error().Msg("File storage is nil")
			return nil, errors.New("file storage is nil")
		}
		if b.cacheStorage == nil {
			b.log.Error().Msg("Cache storage is nil")
			return nil, errors.New("cache storage is nil")
		}
		b.log.Info().Str("func", "MetricsServiceBuilder.buildMetricsService").Msg("CacheMetricsService with File created")
		return &CacheMetricsService{
			cache:         b.cacheStorage,
			file:          b.fileStorage,
			log:           b.log,
			storeInterval: b.cfg.StoreInterval,
		}, nil
	}

	// Cache only
	if b.cacheStorage != nil {
		b.log.Info().Str("func", "MetricsServiceBuilder.buildMetricsService").Msg("CacheMetricsService created")
		return &CacheMetricsService{
			cache: b.cacheStorage,
			log:   b.log,
		}, nil
	}

	b.log.Error().Str("func", "MetricsServiceBuilder.buildMetricsService").Msg("nil storage was provided")
	return nil, errors.New("no valid storage provided")
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
