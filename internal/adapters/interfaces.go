package adapters

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

type AuditEventAdapter interface {
	SendEvent(ctx context.Context, event models.AuditEvent) error
}
