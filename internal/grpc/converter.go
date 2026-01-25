package grpc

import (
	"fmt"

	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/models"
)

type metricConverter struct {
}

func newMetricConverter() MetricsConverter {
	return new(metricConverter)
}

func (c *metricConverter) ConvertMetricFromProto(protoMetric *proto.Metric) (models.Metrics, error) {
	mType := protoMetric.GetType()

	switch mType {
	case proto.Metric_COUNTER:
		return counter(protoMetric)
	case proto.Metric_GAUGE:
		return gauge(protoMetric)
	}

	return models.Metrics{}, fmt.Errorf("%w: %+v", errUnsupportedMetricType, mType)
}

func (c *metricConverter) ConvertMetricsFromProto(protoMetrics ...*proto.Metric) ([]models.Metrics, error) {
	if protoMetrics == nil {
		return nil, errNoMetricsProvided
	}

	metrics := make([]models.Metrics, 0, len(protoMetrics))

	for _, protoMetric := range protoMetrics {
		metric, err := c.ConvertMetricFromProto(protoMetric)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func counter(protoMetric *proto.Metric) (models.Metrics, error) {
	delta := protoMetric.GetDelta()
	id := protoMetric.GetId()

	return models.Metrics{
		ID:    id,
		Delta: &delta,
		MType: models.Counter,
	}, nil
}

func gauge(protoMetric *proto.Metric) (models.Metrics, error) {
	value := protoMetric.GetValue()
	id := protoMetric.GetId()

	return models.Metrics{
		ID:    id,
		Value: &value,
		MType: models.Gauge,
	}, nil
}
