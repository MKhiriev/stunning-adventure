package service

import (
	"context"
	"errors"

	"github.com/MKhiriev/stunning-adventure/internal/adapters"
	"github.com/MKhiriev/stunning-adventure/internal/service/observer"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

// auditService is enabled only when URL and File are provided
// this service is publisher (probably)
type auditService struct {
	observers map[string]observer.AuditObserver
	logger    *zerolog.Logger
}

func NewAuditService(filePath string, adapter adapters.AuditEventAdapter, logger *zerolog.Logger) AuditPublisher {
	service := &auditService{
		observers: make(map[string]observer.AuditObserver),
		logger:    logger,
	}

	observers := observer.NewObservers(filePath, adapter, logger)

	if err := service.Register(observers.FileObserver); err != nil {
		service.logger.Err(err).Str("func", "service.NewAuditService").
			Msg("file observer was not created")
	}

	if err := service.Register(observers.RemoteServerObserver); err != nil {
		service.logger.Err(err).Str("func", "service.NewAuditService").
			Msg("remote server observer was not created")
	}

	return service
}

func (a *auditService) NotifyAll(ctx context.Context, event models.AuditEvent) error {
	var errs []error // = make([]error, len(a.observers))

	for _, auditObserver := range a.observers {
		err := auditObserver.Update(ctx, event)
		if err != nil {
			errs = append(errs, err)
			a.logger.Err(err).
				Str("func", "*auditService.NotifyAll").
				Str("auditObserver", auditObserver.Name()).
				Msg("error occurred updating")
		}
	}

	if len(errs) != 0 {
		return observer.MultiObserverError{Errors: errs}
	}

	return nil
}

func (a *auditService) Register(observer observer.AuditObserver) error {
	if observer == nil {
		a.logger.Error().Str("func", "*auditService.Register").Msg("nil observer was passed")
		return errors.New("cannot register nil observer")
	}

	a.logger.Info().Str("func", "*auditService.Register").
		Str("observer", observer.Name()).Msg("observer registered")
	a.observers[observer.Name()] = observer

	return nil
}

func (a *auditService) Deregister(observer observer.AuditObserver) error {
	if observer == nil {
		a.logger.Error().Str("func", "*auditService.Register").Msg("nil observer was passed")
		return errors.New("cannot register nil observer")
	}

	// check if observer was registered
	if _, ok := a.observers[observer.Name()]; ok {
		a.logger.Info().Str("func", "*auditService.Register").Msg("nil observer was passed")

		// if was registered before
		delete(a.observers, observer.Name())
		return nil
	}

	return errors.New("provided observer wasn't registered")
}
