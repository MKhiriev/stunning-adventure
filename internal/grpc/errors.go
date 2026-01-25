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

	errMissingMetadata       = status.Error(codes.PermissionDenied, "missing metadata")
	errMissingRealIpMetadata = status.Error(codes.PermissionDenied, "missing x-real-ip in metadata")
	errEmptyRealIpMetadata   = status.Error(codes.PermissionDenied, "x-real-ip is empty")

	errInvalidIPAddress     = status.Error(codes.PermissionDenied, "invalid IP address")
	errNotFromTrustedSubnet = status.Error(codes.PermissionDenied, "IP address not in trusted subnet")

	errUnsupportedMetricType = errors.New("unsupported metric type")
)
