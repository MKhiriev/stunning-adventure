package service

import (
	"context"
	"errors"

	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog"
)

type PingDBService struct {
	*store.DB
	log *zerolog.Logger
}

func (p *PingDBService) Ping(ctx context.Context) error {
	if p.DB != nil {
		p.log.Info().Str("func", "*PingDBService.Ping").Msg("Ping to DB sent")
		return p.DB.PingContext(ctx)
	}

	p.log.Error().Str("func", "*PingDBService.Ping").Msg("error during DB ping: db pointer is nil")
	return errors.New("DB connection is nil")
}

func NewPingDBService(db *store.DB, log *zerolog.Logger) (PingService, error) {
	log.Info().Str("func", "service.NewPingDBService").Msg("PingDBService successfully created")
	return &PingDBService{DB: db, log: log}, nil
}
