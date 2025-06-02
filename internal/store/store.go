package store

import "github.com/MKhiriev/stunning-adventure/models"

type MemStorage struct {
	Memory map[string]models.Metrics
}
