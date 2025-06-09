package agent

import (
	"github.com/MKhiriev/stunning-adventure/models"
)

type Agent interface {
	ReadMetrics() error
	SendMetrics() error
	Run() error
}

type MemStorage interface {
	GetAllMetrics() []models.Metrics
	RefreshAllMetrics(metrics ...models.Metrics)
	Flush()
}
