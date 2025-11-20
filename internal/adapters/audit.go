package adapters

import (
	"context"
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type auditAdapter struct {
	remoteServer string
	client       *resty.Client
	logger       *zerolog.Logger
}

func NewAuditAdapter(remoteServer string, logger *zerolog.Logger) AuditEventAdapter {
	if remoteServer == "" {
		return nil
	}

	logger.Debug().Str("func", "adapters.NewAuditAdapter").Msg("audit adapter is initialized")
	return &auditAdapter{
		remoteServer: remoteServer,
		client:       utils.NewHTTPClient(),
		logger:       logger,
	}
}

// SendEvent function performs an audit of the received metrics
func (a *auditAdapter) SendEvent(ctx context.Context, event models.AuditEvent) error {
	response, err := a.client.R().
		SetContext(ctx).
		SetBody(event).
		Post(a.remoteServer)
	if err != nil {
		a.logger.Err(err).Str("func", "*auditAdapter.SendEvent").
			Any("event", event).
			Msg("error occurred sending event")
		return fmt.Errorf("%w: %w", ErrEventNotSent, err)
	}

	a.logger.Debug().Str("func", "*auditAdapter.SendEvent").
		Any("event", event).
		Str("status", response.Status()).
		Bytes("response body", response.Body()).
		Msg("event is sent")

	return nil
}
