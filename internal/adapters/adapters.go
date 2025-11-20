package adapters

import (
	"github.com/rs/zerolog"
)

type Adapters struct {
	AuditEventAdapter AuditEventAdapter
}

func NewAdapters(remoteServer string, logger *zerolog.Logger) *Adapters {
	defer logger.Info().Msg("adapters are initialized")

	return &Adapters{
		AuditEventAdapter: NewAuditAdapter(remoteServer),
	}
}
