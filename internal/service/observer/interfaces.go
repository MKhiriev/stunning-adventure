package observer

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

// AuditObserver defines the interface for components that handle audit events.
// Implementations receive notifications from an AuditPublisher and perform
// side effect actions such as sending data to external systems,
// or saving to a file.
type AuditObserver interface {
	Update(ctx context.Context, event models.AuditEvent) error // processes a single audit event
	Name() string                                              // returns observer name for logging
}
