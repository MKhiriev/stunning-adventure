package models

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
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

func NewMetricWithoutValue(ID, MType string) (Metrics, error) {
	// check if not nil vals and type is Counter or gauge
	if ID == "" || MType == "" || !slices.Contains(allowedTypes, MType) {
		return Metrics{}, errors.New("passed metric params are not valid")
	}

	return Metrics{ID: ID, MType: MType}, nil
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

func (m *Metrics) String() string {
	if m.MType == Gauge {
		return fmt.Sprintf(`{ID: %s, MType: %s, Value: %.0f}`,
			m.ID, m.MType, *m.Value)
	}

	return fmt.Sprintf(`{ID: %s, MType: %s, Delta: %d}`,
		m.ID, m.MType, *m.Delta)
}
