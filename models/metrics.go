package models

import (
	"fmt"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

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

func (m *Metrics) String() string {
	if m.MType == Gauge {
		return fmt.Sprintf(`{ID: %s, MType: %s, Value: %.0f}`,
			m.ID, m.MType, *m.Value)
	}

	return fmt.Sprintf(`{ID: %s, MType: %s, Delta: %d}`,
		m.ID, m.MType, *m.Delta)
}
