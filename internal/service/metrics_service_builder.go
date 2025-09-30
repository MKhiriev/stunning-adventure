package service

import (
	"errors"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog"
)

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

func (b *MetricsServiceBuilder) Build() (MetricsService, error) {
	var service MetricsService

	b.log.Info().Str("func", "MetricsServiceBuilder.Build").Msg("started building MetricsService")
	service, err := b.buildMetricsService()
	if err != nil {
		return nil, fmt.Errorf("error occurred during building metrics service %w", err)
	}

	// adding wrappers if they exist
	if len(b.wrappers) > 0 {
		b.log.Info().Str("func", "MetricsServiceBuilder.Build").Msg("started adding wrapper services for the MetricsService")
		for _, wrapper := range b.wrappers {
			service = wrapper.Wrap(service)
		}
	}

	return service, nil
}

func (b *MetricsServiceBuilder) buildMetricsService() (MetricsService, error) {
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
