// Package models provides core data structures and types used across the metrics system.
// It defines metric types, their serialization formats, and utility functions for creation
// and string representation. Metrics are represented in a flat structure, distinguishing
// between unset and zero values via pointer fields (Delta and Value).
//
// Supported metric types:
//   - Counter: integer-based metrics representing cumulative counts
//   - Gauge: floating-point metrics representing instantaneous measurements
package models

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

const (
	// Counter identifies a metric of type "counter".
	// Counter metrics represent integer increments accumulated over time.
	Counter = "counter"

	// Gauge identifies a metric of type "gauge".
	// Gauge metrics represent floating-point values sampled at a moment in time.
	Gauge = "gauge"
)

var allowedTypes = []string{Counter, Gauge}

// Metrics NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// NewMetric creates a fully-specified metric from textual inputs.
// It validates the metric type, checks for non-empty parameters,
// and parses the textual value according to the metric type.
//
// Behavior:
//   - Returns an error if ID, MType, or value are empty
//   - Rejects unsupported metric types
//   - Parses value as float64 for gauge or int64 for counter
//
// Parameters:
//
//	ID    - metric identifier string
//	MType - metric type ("counter" or "gauge")
//	Value - string value to be parsed into Delta or Value
//
// Returns:
//
//	Metrics - constructed metric with parsed value
//	error   - validation or parsing failure
func NewMetric(ID, MType, Value string) (Metrics, error) {
	// check if not nil vals and type is Counter or gauge
	if ID == "" || MType == "" || !slices.Contains(allowedTypes, MType) || Value == "" {
		return Metrics{}, errors.New("passed metric params are not valid")
	}

	var err error
	var metric Metrics
	switch MType {
	case Gauge:
		metric, err = newGauge(ID, MType, Value)
	case Counter:
		metric, err = newCounter(ID, MType, Value)
	}
	if err != nil {
		return Metrics{}, fmt.Errorf("error occured during mteric creation: %w", err)
	}

	return metric, nil
}

func newGauge(ID, MType, Value string) (Metrics, error) {
	gaugeValue, conversionError := strconv.ParseFloat(Value, 64)
	if conversionError != nil {
		return Metrics{}, errors.New("passed GAUGE metric params are not valid")
	}

	return Metrics{
		ID:    ID,
		MType: MType,
		Value: &gaugeValue,
	}, nil
}

func newCounter(ID, MType, Value string) (Metrics, error) {
	counterValue, conversionError := strconv.ParseInt(Value, 10, 64)
	if conversionError != nil {
		return Metrics{}, errors.New("passed COUNTER metric params are not valid")
	}

	return Metrics{
		ID:    ID,
		MType: MType,
		Delta: &counterValue,
	}, nil
}

// String returns a human-readable string representation of a metric,
// selecting the appropriate value field depending on its type.
func (m *Metrics) String() string {
	if m.MType == Gauge {
		return fmt.Sprintf(`{ID: %s, MType: %s, Value: %.0f}`,
			m.ID, m.MType, *m.Value)
	}

	return fmt.Sprintf(`{ID: %s, MType: %s, Delta: %d}`,
		m.ID, m.MType, *m.Delta)
}
