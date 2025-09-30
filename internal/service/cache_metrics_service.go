package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type CacheMetricsService struct {
	file          store.Storage // secondary chosen store.Storage provider - file
	cache         store.Storage // primary chosen store.Storage provider - cache
	storeInterval int64         // for file-storage. Only needed if file storage is provided
	log           *zerolog.Logger
}

func (c *CacheMetricsService) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	// save metric in cache memory
	result, err := c.cache.Save(ctx, metric)
	if err != nil {
		c.log.Err(err).Str("func", "*CacheMetricsService.Save").Msg("error during saving metric in cache")
		return models.Metrics{}, fmt.Errorf("error during saving metric in cache: %w", err)
	}

	// if file storage was provided
	c.log.Debug().Str("func", "*CacheMetricsService.Save").Bool("file exists?", c.file != nil).Msg("checking if file exists")
	if c.file != nil {
		// get all metrics from Cache storage
		allMetrics, err := c.cache.GetAll(ctx)
		if err != nil {
			c.log.Err(err).Str("func", "*CacheMetricsService.Save").Msg("error during getting all metrics from cache")
			return models.Metrics{}, fmt.Errorf("error during getting all metrics from cache: %w", err)
		}

		// save all metrics to file
		err = c.file.SaveAll(ctx, allMetrics)
		if err != nil {
			c.log.Err(err).Str("func", "*CacheMetricsService.Save").Msg("error during saving all metrics to file")
			return models.Metrics{}, fmt.Errorf("error during saving all metrics to file: %w", err)
		}
	}

	return result, nil
}

func (c *CacheMetricsService) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	// save all metrics in cache memory
	err := c.cache.SaveAll(ctx, metrics)
	if err != nil {
		c.log.Err(err).Str("func", "*CacheMetricsService.SaveAll").Msg("error during saving all metrics in cache")
		return fmt.Errorf("error during saving all metrics in cache: %w", err)
	}

	// if file storage was provided
	c.log.Debug().Str("func", "*CacheMetricsService.SaveAll").Bool("file exists?", c.file != nil).Msg("checking if file exists")
	if c.file != nil {
		// save all provided metrics
		err = c.file.SaveAll(ctx, metrics)
		if err != nil {
			c.log.Err(err).Str("func", "*CacheMetricsService.SaveAll").Msg("error during saving all metrics to file")
			return fmt.Errorf("error during saving all metrics to file: %w", err)
		}
	}

	return nil
}

func (c *CacheMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return c.cache.Get(ctx, metric)
}

func (c *CacheMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return c.cache.GetAll(ctx)
}
