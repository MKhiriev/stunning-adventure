package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetric_Success_Gauge(t *testing.T) {
	m, err := NewMetric("temperature", Gauge, "12.34")
	require.NoError(t, err)

	assert.Equal(t, "temperature", m.ID)
	assert.Equal(t, Gauge, m.MType)
	require.NotNil(t, m.Value)
	assert.InDelta(t, 12.34, *m.Value, 1e-9)

	assert.Nil(t, m.Delta, "gauge must not set Delta")
}

func TestNewMetric_Success_Counter(t *testing.T) {
	m, err := NewMetric("requests", Counter, "42")
	require.NoError(t, err)

	assert.Equal(t, "requests", m.ID)
	assert.Equal(t, Counter, m.MType)
	require.NotNil(t, m.Delta)
	assert.EqualValues(t, 42, *m.Delta)

	assert.Nil(t, m.Value, "counter must not set Value")
}

func TestNewMetric_Rejects_EmptyFields(t *testing.T) {
	cases := []struct {
		name  string
		id    string
		mtype string
		val   string
	}{
		{"empty id", "", Gauge, "1.0"},
		{"empty type", "m1", "", "1.0"},
		{"empty value", "m1", Gauge, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := NewMetric(tc.id, tc.mtype, tc.val)
			require.Error(t, err)
			assert.Empty(t, m.ID)
			assert.Contains(t, err.Error(), "passed metric params are not valid")
		})
	}
}

func TestNewMetric_Rejects_UnsupportedType(t *testing.T) {
	m, err := NewMetric("m1", "histogram", "10")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "passed metric params are not valid")
	assert.Equal(t, Metrics{}, m)
}

func TestNewMetric_ParseError_Gauge(t *testing.T) {
	m, err := NewMetric("m1", Gauge, "not-a-float")
	require.Error(t, err)

	// outer wrapper text
	assert.Contains(t, err.Error(), "error occured during mteric creation")
	// inner error text
	assert.Contains(t, err.Error(), "passed GAUGE metric params are not valid")

	assert.Equal(t, Metrics{}, m)
}

func TestNewMetric_ParseError_Counter(t *testing.T) {
	m, err := NewMetric("m1", Counter, "not-an-int")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error occured during mteric creation")
	assert.Contains(t, err.Error(), "passed COUNTER metric params are not valid")

	assert.Equal(t, Metrics{}, m)
}

func TestNewMetric_Allows_NegativeValues(t *testing.T) {
	t.Run("counter negative", func(t *testing.T) {
		m, err := NewMetric("m1", Counter, "-5")
		require.NoError(t, err)
		require.NotNil(t, m.Delta)
		assert.EqualValues(t, -5, *m.Delta)
	})

	t.Run("gauge negative", func(t *testing.T) {
		m, err := NewMetric("m1", Gauge, "-3.14")
		require.NoError(t, err)
		require.NotNil(t, m.Value)
		assert.InDelta(t, -3.14, *m.Value, 1e-9)
	})
}

func TestMetrics_String_Counter(t *testing.T) {
	v := int64(7)
	m := &Metrics{
		ID:    "hits",
		MType: Counter,
		Delta: &v,
	}

	assert.Equal(t, `{ID: hits, MType: counter, Delta: 7}`, m.String())
}

func TestMetrics_String_Gauge_RoundsToZeroDecimals(t *testing.T) {
	v := 12.34
	m := &Metrics{
		ID:    "temp",
		MType: Gauge,
		Value: &v,
	}

	// String() uses "%.0f" -> rounds to nearest integer (12.34 -> 12)
	assert.Equal(t, `{ID: temp, MType: gauge, Value: 12}`, m.String())
}

func TestMetrics_String_Gauge_RoundingBehavior(t *testing.T) {
	v := 12.6
	m := &Metrics{
		ID:    "temp",
		MType: Gauge,
		Value: &v,
	}

	// 12.6 -> 13 with "%.0f"
	assert.Equal(t, `{ID: temp, MType: gauge, Value: 13}`, m.String())
}
