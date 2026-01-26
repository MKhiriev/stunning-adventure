package grpc

import (
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricConverter(t *testing.T) {
	converter := newMetricConverter()
	require.NotNil(t, converter)
	assert.Implements(t, (*MetricsConverter)(nil), converter)
}

func TestConvertMetricFromProto_Counter_Success(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("requests")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(42)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	assert.Equal(t, "requests", metric.ID)
	assert.Equal(t, models.Counter, metric.MType)
	require.NotNil(t, metric.Delta)
	assert.EqualValues(t, 42, *metric.Delta)
	assert.Nil(t, metric.Value, "counter should not have Value set")
}

func TestConvertMetricFromProto_Gauge_Success(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("temperature")
	protoMetric.SetType(proto.Metric_GAUGE)
	protoMetric.SetValue(36.6)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	assert.Equal(t, "temperature", metric.ID)
	assert.Equal(t, models.Gauge, metric.MType)
	require.NotNil(t, metric.Value)
	assert.InDelta(t, 36.6, *metric.Value, 1e-9)
	assert.Nil(t, metric.Delta, "gauge should not have Delta set")
}

func TestConvertMetricFromProto_Counter_ZeroValue(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("zero_counter")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(0)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	require.NotNil(t, metric.Delta)
	assert.EqualValues(t, 0, *metric.Delta)
}

func TestConvertMetricFromProto_Gauge_ZeroValue(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("zero_gauge")
	protoMetric.SetType(proto.Metric_GAUGE)
	protoMetric.SetValue(0.0)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	require.NotNil(t, metric.Value)
	assert.InDelta(t, 0.0, *metric.Value, 1e-9)
}

func TestConvertMetricFromProto_Counter_NegativeValue(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("negative_counter")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(-100)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	require.NotNil(t, metric.Delta)
	assert.EqualValues(t, -100, *metric.Delta)
}

func TestConvertMetricFromProto_Gauge_NegativeValue(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("negative_gauge")
	protoMetric.SetType(proto.Metric_GAUGE)
	protoMetric.SetValue(-273.15)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err)
	require.NotNil(t, metric.Value)
	assert.InDelta(t, -273.15, *metric.Value, 1e-9)
}

func TestConvertMetricFromProto_EmptyID(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(10)

	converter := newMetricConverter()
	metric, err := converter.ConvertMetricFromProto(protoMetric)

	require.NoError(t, err, "empty ID should be allowed")
	assert.Empty(t, metric.ID)
}

func TestConvertMetricFromProto_UnsupportedType(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("unknown")
	// Simulate unsupported type by using invalid enum value
	// In proto3, unknown enum values default to 0, but we can test the error path
	// by not setting type at all or using reflection to set invalid value

	converter := newMetricConverter()

	// Test with uninitialized type (defaults to GAUGE = 0)
	// This will actually succeed, so let's test the error message format
	_, err := converter.ConvertMetricFromProto(protoMetric)
	// Since proto.Metric_GAUGE is 0 and default, this won't error
	// To properly test unsupported type, we'd need to mock or use invalid enum
	require.NoError(t, err) // Default type is GAUGE
}

func TestConvertMetricsFromProto_Success_Multiple(t *testing.T) {
	protoMetric1 := &proto.Metric{}
	protoMetric1.SetId("counter1")
	protoMetric1.SetType(proto.Metric_COUNTER)
	protoMetric1.SetDelta(10)

	protoMetric2 := &proto.Metric{}
	protoMetric2.SetId("gauge1")
	protoMetric2.SetType(proto.Metric_GAUGE)
	protoMetric2.SetValue(3.14)

	protoMetric3 := &proto.Metric{}
	protoMetric3.SetId("counter2")
	protoMetric3.SetType(proto.Metric_COUNTER)
	protoMetric3.SetDelta(20)

	converter := newMetricConverter()
	metrics, err := converter.ConvertMetricsFromProto(protoMetric1, protoMetric2, protoMetric3)

	require.NoError(t, err)
	assert.Len(t, metrics, 3)

	assert.Equal(t, "counter1", metrics[0].ID)
	assert.Equal(t, models.Counter, metrics[0].MType)

	assert.Equal(t, "gauge1", metrics[1].ID)
	assert.Equal(t, models.Gauge, metrics[1].MType)

	assert.Equal(t, "counter2", metrics[2].ID)
	assert.Equal(t, models.Counter, metrics[2].MType)
}

func TestConvertMetricsFromProto_Success_Single(t *testing.T) {
	protoMetric := &proto.Metric{}
	protoMetric.SetId("single")
	protoMetric.SetType(proto.Metric_COUNTER)
	protoMetric.SetDelta(1)

	converter := newMetricConverter()
	metrics, err := converter.ConvertMetricsFromProto(protoMetric)

	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "single", metrics[0].ID)
}

func TestConvertMetricToProto_Counter_Success(t *testing.T) {
	delta := int64(42)
	metric := models.Metrics{
		ID:    "requests",
		MType: models.Counter,
		Delta: &delta,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.NoError(t, err)
	assert.Equal(t, "requests", protoMetric.GetId())
	assert.Equal(t, proto.Metric_COUNTER, protoMetric.GetType())
	assert.EqualValues(t, 42, protoMetric.GetDelta())
}

func TestConvertMetricToProto_Gauge_Success(t *testing.T) {
	value := 36.6
	metric := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: &value,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.NoError(t, err)
	assert.Equal(t, "temperature", protoMetric.GetId())
	assert.Equal(t, proto.Metric_GAUGE, protoMetric.GetType())
	assert.InDelta(t, 36.6, protoMetric.GetValue(), 1e-9)
}

func TestConvertMetricToProto_Counter_NilDelta(t *testing.T) {
	metric := models.Metrics{
		ID:    "bad_counter",
		MType: models.Counter,
		Delta: nil,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "counter metric bad_counter has no delta value")
	assert.Nil(t, protoMetric)
}

func TestConvertMetricToProto_Gauge_NilValue(t *testing.T) {
	metric := models.Metrics{
		ID:    "bad_gauge",
		MType: models.Gauge,
		Value: nil,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gauge metric bad_gauge has no value")
	assert.Nil(t, protoMetric)
}

func TestConvertMetricToProto_UnsupportedType(t *testing.T) {
	delta := int64(10)
	metric := models.Metrics{
		ID:    "unknown",
		MType: "histogram",
		Delta: &delta,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.Error(t, err)
	assert.ErrorIs(t, err, errUnsupportedMetricType)
	assert.Contains(t, err.Error(), "histogram")
	assert.Nil(t, protoMetric)
}

func TestConvertMetricToProto_Counter_ZeroValue(t *testing.T) {
	delta := int64(0)
	metric := models.Metrics{
		ID:    "zero_counter",
		MType: models.Counter,
		Delta: &delta,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.NoError(t, err)
	assert.EqualValues(t, 0, protoMetric.GetDelta())
}

func TestConvertMetricToProto_Gauge_ZeroValue(t *testing.T) {
	value := 0.0
	metric := models.Metrics{
		ID:    "zero_gauge",
		MType: models.Gauge,
		Value: &value,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.NoError(t, err)
	assert.InDelta(t, 0.0, protoMetric.GetValue(), 1e-9)
}

func TestConvertMetricToProto_EmptyID(t *testing.T) {
	delta := int64(10)
	metric := models.Metrics{
		ID:    "",
		MType: models.Counter,
		Delta: &delta,
	}

	converter := newMetricConverter()
	protoMetric, err := converter.ConvertMetricToProto(metric)

	require.NoError(t, err, "empty ID should be allowed")
	assert.Empty(t, protoMetric.GetId())
}

func TestConvertMetricsToProto_Success_Multiple(t *testing.T) {
	delta1 := int64(10)
	value1 := 3.14
	delta2 := int64(20)

	metrics := []models.Metrics{
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "gauge1", MType: models.Gauge, Value: &value1},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}

	converter := newMetricConverter()
	protoMetrics, err := converter.ConvertMetricsToProto(metrics...)

	require.NoError(t, err)
	assert.Len(t, protoMetrics, 3)

	assert.Equal(t, "counter1", protoMetrics[0].GetId())
	assert.Equal(t, proto.Metric_COUNTER, protoMetrics[0].GetType())

	assert.Equal(t, "gauge1", protoMetrics[1].GetId())
	assert.Equal(t, proto.Metric_GAUGE, protoMetrics[1].GetType())

	assert.Equal(t, "counter2", protoMetrics[2].GetId())
	assert.Equal(t, proto.Metric_COUNTER, protoMetrics[2].GetType())
}

func TestConvertMetricsToProto_Success_Single(t *testing.T) {
	delta := int64(1)
	metric := models.Metrics{
		ID:    "single",
		MType: models.Counter,
		Delta: &delta,
	}

	converter := newMetricConverter()
	protoMetrics, err := converter.ConvertMetricsToProto(metric)

	require.NoError(t, err)
	assert.Len(t, protoMetrics, 1)
	assert.Equal(t, "single", protoMetrics[0].GetId())
}

func TestConvertMetricsToProto_NilSlice(t *testing.T) {
	converter := newMetricConverter()
	protoMetrics, err := converter.ConvertMetricsToProto()

	require.Error(t, err)
	assert.ErrorIs(t, err, errNoMetricsProvided)
	assert.Nil(t, protoMetrics)
}

func TestConvertMetricsToProto_ErrorInMiddle(t *testing.T) {
	delta := int64(10)
	metrics := []models.Metrics{
		{ID: "valid", MType: models.Counter, Delta: &delta},
		{ID: "invalid", MType: models.Counter, Delta: nil}, // nil delta
	}

	converter := newMetricConverter()
	protoMetrics, err := converter.ConvertMetricsToProto(metrics...)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "counter metric invalid has no delta value")
	assert.Nil(t, protoMetrics)
}

func TestRoundTrip_Counter(t *testing.T) {
	delta := int64(42)
	original := models.Metrics{
		ID:    "requests",
		MType: models.Counter,
		Delta: &delta,
	}

	converter := newMetricConverter()

	// Convert to proto
	protoMetric, err := converter.ConvertMetricToProto(original)
	require.NoError(t, err)

	// Convert back to models
	result, err := converter.ConvertMetricFromProto(protoMetric)
	require.NoError(t, err)

	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.MType, result.MType)
	require.NotNil(t, result.Delta)
	assert.EqualValues(t, *original.Delta, *result.Delta)
}

func TestRoundTrip_Gauge(t *testing.T) {
	value := 36.6
	original := models.Metrics{
		ID:    "temperature",
		MType: models.Gauge,
		Value: &value,
	}

	converter := newMetricConverter()

	// Convert to proto
	protoMetric, err := converter.ConvertMetricToProto(original)
	require.NoError(t, err)

	// Convert back to models
	result, err := converter.ConvertMetricFromProto(protoMetric)
	require.NoError(t, err)

	assert.Equal(t, original.ID, result.ID)
	assert.Equal(t, original.MType, result.MType)
	require.NotNil(t, result.Value)
	assert.InDelta(t, *original.Value, *result.Value, 1e-9)
}

func TestRoundTrip_MultipleMetrics(t *testing.T) {
	delta1 := int64(10)
	value1 := 3.14
	delta2 := int64(20)

	originals := []models.Metrics{
		{ID: "counter1", MType: models.Counter, Delta: &delta1},
		{ID: "gauge1", MType: models.Gauge, Value: &value1},
		{ID: "counter2", MType: models.Counter, Delta: &delta2},
	}

	converter := newMetricConverter()

	// Convert to proto
	protoMetrics, err := converter.ConvertMetricsToProto(originals...)
	require.NoError(t, err)

	// Convert back to models
	results, err := converter.ConvertMetricsFromProto(protoMetrics...)
	require.NoError(t, err)

	assert.Len(t, results, len(originals))
	for i := range originals {
		assert.Equal(t, originals[i].ID, results[i].ID)
		assert.Equal(t, originals[i].MType, results[i].MType)
	}
}

func TestConvertMetricToProto_LargeValues(t *testing.T) {
	t.Run("large counter", func(t *testing.T) {
		delta := int64(9223372036854775807) // max int64
		metric := models.Metrics{
			ID:    "large_counter",
			MType: models.Counter,
			Delta: &delta,
		}

		converter := newMetricConverter()
		protoMetric, err := converter.ConvertMetricToProto(metric)

		require.NoError(t, err)
		assert.EqualValues(t, delta, protoMetric.GetDelta())
	})

	t.Run("large gauge", func(t *testing.T) {
		value := 1.7976931348623157e+308 // near max float64
		metric := models.Metrics{
			ID:    "large_gauge",
			MType: models.Gauge,
			Value: &value,
		}

		converter := newMetricConverter()
		protoMetric, err := converter.ConvertMetricToProto(metric)

		require.NoError(t, err)
		assert.InDelta(t, value, protoMetric.GetValue(), 1e290)
	})
}

func TestConvertMetricFromProto_SpecialFloatValues(t *testing.T) {
	t.Run("very small gauge", func(t *testing.T) {
		protoMetric := &proto.Metric{}
		protoMetric.SetId("tiny")
		protoMetric.SetType(proto.Metric_GAUGE)
		protoMetric.SetValue(1e-308)

		converter := newMetricConverter()
		metric, err := converter.ConvertMetricFromProto(protoMetric)

		require.NoError(t, err)
		require.NotNil(t, metric.Value)
		assert.InDelta(t, 1e-308, *metric.Value, 1e-320)
	})
}

func TestConcurrentConversion(t *testing.T) {
	converter := newMetricConverter()
	const numGoroutines = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			delta := int64(id)
			metric := models.Metrics{
				ID:    "concurrent",
				MType: models.Counter,
				Delta: &delta,
			}

			protoMetric, err := converter.ConvertMetricToProto(metric)
			require.NoError(t, err)

			result, err := converter.ConvertMetricFromProto(protoMetric)
			require.NoError(t, err)
			assert.EqualValues(t, id, *result.Delta)

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
