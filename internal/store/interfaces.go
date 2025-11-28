// Package store provides abstractions and interfaces for metric storage layers.
// It defines multiple persistence strategies including in-memory cache, file-based storage,
// and database-backed storage. This package allows flexible swapping of storage
// implementations while keeping the business logic and handlers storage-agnostic.
//
// Core concepts:
//   - Storage: unified interface for CRUD operations on metrics.
//   - MetricsFileStorage: file-based persistence of metrics.
//   - MetricsCacheStorage: fast in-memory storage for counters and gauges.
//   - MetricsDatabaseStorage: database migration and schema operations.
//   - ErrorClassificator: classification of errors for consistent handling across layers.
//
// Metrics are represented by the models.Metrics struct, which can hold
// either a Counter (Delta) or a Gauge (Value). Storage implementations
// are expected to handle the specific semantics for each metric type.
//
// Typical usage involves:
//  1. Creating a storage implementation (memory, file, database).
//  2. Wrapping the storage behind the Storage interface.
//  3. Using the Storage methods to Save, Get, or retrieve all metrics.
//  4. Optionally persisting to file or migrating a database schema.
package store

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

// Storage defines the primary interface for metric persistence layers.
// It abstracts interactions with any underlying storage engine
// (file, memory cache, database, or hybrid).
type Storage interface {
	Save(context.Context, models.Metrics) (models.Metrics, error) // persists a single metric and returns the updated stored version
	SaveAll(context.Context, []models.Metrics) error              // persists a batch of metrics atomically
	Get(context.Context, models.Metrics) (models.Metrics, error)  // retrieves a single metric by its identifier and type
	GetAll(context.Context) ([]models.Metrics, error)             // returns all stored metrics
}

// MetricsFileStorage defines an interface for handling metric persistence
// backed by file-based storage systems.
type MetricsFileStorage interface {
	SaveMetricsToFile(context.Context, []models.Metrics) error     // writes all metrics to a storage file
	LoadMetricsFromFile(context.Context) ([]models.Metrics, error) // restores metrics from file on startup
}

// MetricsCacheStorage defines an in-memory layer for metric manipulation.
// Typically used for fast, non-persistent operations before flushing to disk/database.
type MetricsCacheStorage interface {
	AddCounter(context.Context, models.Metrics) (models.Metrics, error)                                      // increments or creates a counter metric
	UpdateGauge(context.Context, models.Metrics) (models.Metrics, error)                                     // sets or updates a gauge metric
	GetMetricByNameAndType(ctx context.Context, metricName string, metricType string) (models.Metrics, bool) // retrieves a metric by ID and type, returning (metric, found)
	GetAllMetrics(context.Context) []models.Metrics                                                          // returns all metrics from cache memory
}

// MetricsDatabaseStorage defines persistence logic for database-backed storage layers.
type MetricsDatabaseStorage interface {
	Migrate(context.Context) error // performs schema setup or migrations required for the metrics table
}

// ErrorClassificator defines a strategy for categorizing errors produced by persistence layers.
type ErrorClassificator interface {
	Classify(err error) ErrorClassification // maps an error into a predefined ErrorClassification enum
}
