// Package agent provides abstractions for the local metric collection agent.
// The agent runs on a host machine, collects metrics, caches them in memory,
// and periodically sends them to a central server.
//
// Core concepts:
//   - Agent: defines the lifecycle and operations of a metric-collecting agent,
//     including reading, sending, and running metrics collection routines.
//   - MemStorage: defines an in-memory storage interface for temporarily holding
//     collected metrics before dispatch to the server.
//
// Usage patterns:
//  1. Implement Agent to define host-specific metric collection and sending logic.
//  2. Implement MemStorage to maintain a transient, memory-based cache of metrics.
//  3. Combine Agent and MemStorage to efficiently collect, buffer, and transmit metrics.
//
// This package decouples metric collection logic from transport and persistence layers,
// enabling modular, testable, and extendable agent implementations.
package agent

import (
	"context"

	"github.com/MKhiriev/stunning-adventure/models"
)

// Agent defines the contract for a local metrics collection agent.
type Agent interface {
	ReadMetrics() error // collects metrics from the host environment
	SendMetrics() error // transmits collected metrics to a server
	Run() error         // starts the agent lifecycle including collection and sending
}

// MemStorage defines the in-memory cache for metrics collected by an agent.
type MemStorage interface {
	GetAllMetrics() []models.Metrics             // retrieves all cached metrics
	RefreshAllMetrics(metrics ...models.Metrics) // updates or adds metrics in memory
	Flush()                                      // clears all cached metrics
}

// Client defines a client that sends metrics to a server.
type Client interface {
	Send(ctx context.Context, metrics ...models.Metrics) error
	Close() error
}
