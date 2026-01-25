package grpc

import (
	"github.com/MKhiriev/stunning-adventure/internal/proto"
	"github.com/MKhiriev/stunning-adventure/models"
)

type MetricsConverter interface {
	ConvertMetricFromProto(protoMetric *proto.Metric) (models.Metrics, error)
	ConvertMetricsFromProto(protoMetrics ...*proto.Metric) ([]models.Metrics, error)
}
