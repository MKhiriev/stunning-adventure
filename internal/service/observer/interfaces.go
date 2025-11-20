package observer

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

type AuditObserver interface {
	Update(ctx context.Context, event models.AuditEvent) error
	Name() string
}
