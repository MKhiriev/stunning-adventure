package validators

import "errors"

var (
	ErrUnsupportedType = errors.New("unsupported type for validation")
	ErrEmptyID         = errors.New("metric name (id) is empty")
	ErrEmptyType       = errors.New("metric type is empty")
	ErrInvalidType     = errors.New("metric type is not valid")
	ErrNoValue         = errors.New("metric has no value")
	ErrUnknownField    = errors.New("unknown field for validation")
)
