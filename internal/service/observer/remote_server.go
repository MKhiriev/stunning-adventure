package observer

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/internal/adapters"
	"github.com/MKhiriev/stunning-adventure/models"
)

type RemoteServerObserver struct {
	auditAdapter adapters.AuditEventAdapter
	description  string
}

func NewRemoteServerObserver(adapter adapters.AuditEventAdapter) AuditObserver {
	return &RemoteServerObserver{auditAdapter: adapter, description: "remote server observer"}
}

func (r *RemoteServerObserver) Update(ctx context.Context, event models.AuditEvent) error {
	//TODO implement me
	panic("implement me")
}

func (r *RemoteServerObserver) Name() string {
	return r.description
}
