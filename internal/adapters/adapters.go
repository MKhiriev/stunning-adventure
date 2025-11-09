package adapters

import (
	"github.com/rs/zerolog"
)

type Adapters struct {
	AuditAdapter
}

func NewAdapters(logger *zerolog.Logger) *Adapters {
	defer logger.Info().Msg("adapters are initialized")

	return &Adapters{
		AuditAdapter: NewAuditAdapter(),
	}
}
