package agent

import (
	"github.com/MKhiriev/stunning-adventure/models"
	"maps"
	"slices"
)

type AgentStorage struct {
	Metrics map[string]models.Metrics
}

func NewStorage() *AgentStorage {
	return &AgentStorage{
		Metrics: make(map[string]models.Metrics),
	}
}

func (c *AgentStorage) GetAllMetrics() []models.Metrics {
	return slices.Collect(maps.Values(c.Metrics))
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
		c.Metrics[metric.ID] = metric
	}
}

// Flush method is emptying the agent memory
func (c *AgentStorage) Flush() {
	c.Metrics = make(map[string]models.Metrics)
}

func (c *AgentStorage) GetMetric(metricId string) models.Metrics {
	return models.Metrics{}
}
