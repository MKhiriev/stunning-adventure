package observer

import (
	"github.com/MKhiriev/stunning-adventure/internal/adapters"
)

type Observers struct {
	FileObserver         AuditObserver
	RemoteServerObserver AuditObserver
}

// DONE!
func NewObservers(filePath string, adapter adapters.AuditEventAdapter) Observers {
	observers := Observers{}

	if filePath != "" {
		observers.FileObserver = NewFileObserver(filePath)
	}

	if adapter != nil {
		observers.RemoteServerObserver = NewRemoteServerObserver(adapter)
	}

	return observers
}
