package observer

import (
	"github.com/MKhiriev/stunning-adventure/internal/adapters"
	"github.com/rs/zerolog"
)

type Observers struct {
	FileObserver         AuditObserver
	RemoteServerObserver AuditObserver
}

func NewObservers(filePath string, adapter adapters.AuditEventAdapter, logger *zerolog.Logger) Observers {
	observers := Observers{}

	if filePath != "" {
		observers.FileObserver = NewFileObserver(filePath, logger)
	}

	if adapter != nil {
		observers.RemoteServerObserver = NewRemoteServerObserver(adapter)
	}

	return observers
}
