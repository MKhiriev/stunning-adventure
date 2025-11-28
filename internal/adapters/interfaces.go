// Package adapters provides abstraction layers for integrating with
// external systems, protocols, or services without coupling business logic
// directly to implementation details.
//
// Core concepts:
//   - AuditEventAdapter: defines an interface to send audit events to
//     external systems or services.
//
// Usage patterns:
//  1. Implement AuditEventAdapter to integrate with logging systems, message queues,
//     HTTP APIs, or other audit destinations.
//  2. Inject the adapter into AuditPublisher implementations to forward events.
//
// This package enables decoupling of core service logic from external dependencies,
// promoting testability and flexibility in system integration.
package adapters

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

// AuditEventAdapter defines the contract for sending audit events to
// external systems or services.
type AuditEventAdapter interface {
	SendEvent(ctx context.Context, event models.AuditEvent) error // sends a single audit event
}
