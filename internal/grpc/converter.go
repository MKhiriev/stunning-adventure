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

func (c *metricConverter) ConvertMetricToProto(metric models.Metrics) (*proto.Metric, error) {
	switch metric.MType {
	case models.Counter:
		return counterToProto(metric)
	case models.Gauge:
		return gaugeToProto(metric)
	}

	return nil, fmt.Errorf("%w: %s", errUnsupportedMetricType, metric.MType)
}

func (c *metricConverter) ConvertMetricsToProto(metrics ...models.Metrics) ([]*proto.Metric, error) {
	if metrics == nil {
		return nil, errNoMetricsProvided
	}

	protoMetrics := make([]*proto.Metric, 0, len(metrics))

	for _, metric := range metrics {
		protoMetric, err := c.ConvertMetricToProto(metric)
		if err != nil {
			return nil, err
		}

		protoMetrics = append(protoMetrics, protoMetric)
	}

	return protoMetrics, nil
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

func counterToProto(metric models.Metrics) (*proto.Metric, error) {
	if metric.Delta == nil {
		return nil, fmt.Errorf("counter metric %s has no delta value", metric.ID)
	}

	m := &proto.Metric{}
	m.SetId(metric.ID)
	m.SetType(proto.Metric_COUNTER)
	m.SetDelta(*metric.Delta)
	return m, nil
}

func gaugeToProto(metric models.Metrics) (*proto.Metric, error) {
	if metric.Value == nil {
		return nil, fmt.Errorf("gauge metric %s has no value", metric.ID)
	}

	m := &proto.Metric{}
	m.SetId(metric.ID)
	m.SetType(proto.Metric_GAUGE)
	m.SetValue(*metric.Value)
	return m, nil
}
