package agent

import (
	"maps"
	"slices"

	"github.com/MKhiriev/stunning-adventure/models"
)

type AgentStorage struct {
	metrics map[string]models.Metrics
}

func NewStorage() *AgentStorage {
	return &AgentStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (c *AgentStorage) GetAllMetrics() []models.Metrics {
	return slices.Collect(maps.Values(c.metrics))
}

func (c *AgentStorage) RefreshAllMetrics(metrics ...models.Metrics) {
	// if no metrics are passed (nothing to update) then stop the method
	if len(metrics) == 0 {
		return
	}

	// empty the memory
	c.Flush()

	// assign passed metrics to new map
	for _, metric := range metrics {
		c.metrics[metric.ID] = metric
	}
}

// Flush method is emptying the agent memory
func (c *AgentStorage) Flush() {
	c.metrics = make(map[string]models.Metrics)
}
