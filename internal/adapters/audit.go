package adapters

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

type auditAdapter struct {
	remoteServer string
	// todo implement structure with proper fields
}

func NewAuditAdapter(remoteServer string) AuditEventAdapter {
	// TODO create http client and static request with replaceable body

	return &auditAdapter{
		remoteServer: remoteServer,
	}
}

// SendEvent function performs an audit of the received metrics
func (a *auditAdapter) SendEvent(ctx context.Context, event models.AuditEvent) error {
	//TODO implement me
	return nil
}
