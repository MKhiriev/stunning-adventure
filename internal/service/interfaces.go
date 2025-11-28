// Package service provides abstractions for business logic operations
// related to metrics, health checks, and audit event propagation.
//
// Core concepts:
//   - MetricsService: defines CRUD operations for metrics with support for batch operations.
//   - PingService: provides readiness and liveness checks for dependent services or storage.
//   - MetricsServiceWrapper: enables decoration of MetricsService with additional behavior such as logging or validation.
//   - AuditPublisher: implements the observer pattern to broadcast audit events to registered observers.
//
// Usage patterns:
//  1. Implement MetricsService to persist metrics in chosen storage layers.
//  2. Wrap MetricsService with MetricsServiceWrapper to inject cross-cutting concerns.
//  3. Register AuditObservers to receive events via AuditPublisher.
//  4. Use PingService to perform health-checks on startup or during runtime.
//
// This package separates business service logic from HTTP handlers and storage,
// ensuring decoupled, testable, and composable service layers.
package service

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/internal/service/observer"
	"github.com/MKhiriev/stunning-adventure/models"
)

// MetricsService defines the fundamental operations for metric persistence
// and retrieval. Implementations provide storage-backed behavior.
type MetricsService interface {
	Save(context.Context, models.Metrics) (models.Metrics, error) // persists a single metric and returns the stored value
	SaveAll(context.Context, []models.Metrics) error              // persists a batch of metrics atomically where supported
	Get(context.Context, models.Metrics) (models.Metrics, error)  // retrieves a specific metric by identifier
	GetAll(context.Context) ([]models.Metrics, error)             // returns all stored metrics
}

// PingService defines the minimal health-check contract for storage or
// infrastructure components. Implementations typically execute a readiness
// check such as a database ping.
type PingService interface {
	Ping(ctx context.Context) error
}

// MetricsServiceWrapper defines middleware composition for MetricsService.
// Implementations wrap an existing MetricsService to add behavior such as
// logging or validating.
type MetricsServiceWrapper interface {
	Wrap(MetricsService) MetricsService // returns a decorated MetricsService applying additional behavior
}

// AuditPublisher defines the observer pattern for audit events.
// Observers are registered and deregistered dynamically. NotifyAll broadcasts
// an audit event to all registered observers.
type AuditPublisher interface {
	Register(observer observer.AuditObserver) error               // attaches an AuditObserver
	Deregister(observer observer.AuditObserver) error             // detaches an AuditObserver
	NotifyAll(ctx context.Context, event models.AuditEvent) error // emits an audit event to all observers
}
