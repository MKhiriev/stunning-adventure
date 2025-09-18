package service

import (
	"context"

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

func (c *CacheMetricsService) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	return c.cache.Get(ctx, metric)
}

func (c *CacheMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	return c.cache.GetAll(ctx)
}
