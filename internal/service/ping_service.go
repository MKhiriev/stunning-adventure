package service

import (
	"context"
	"github.com/MKhiriev/stunning-adventure/internal/store"
)

type PingDBService struct {
	*store.DB
}

func (p *PingDBService) Ping(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

func NewPingDBService(db *store.DB) PingService {
	return &PingDBService{DB: db}
}
