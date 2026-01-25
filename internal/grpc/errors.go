package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errNilRequest        = status.Error(codes.InvalidArgument, "nil request was sent")
	errNoMetricsProvided = status.Error(codes.InvalidArgument, "no metrics to update provided")
	errInvalidMetrics    = status.Error(codes.InvalidArgument, "invalid metric(s)")
	errUnexpectedError   = status.Error(codes.Unknown, "invalid metric(s)")

	errUnsupportedMetricType = errors.New("unsupported metric type")
)
