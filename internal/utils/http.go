package utils

import (
	"time"

	"github.com/go-resty/resty/v2"
)

// NewHTTPClient constructs and returns a Resty HTTP client with debugging disabled.
func NewHTTPClient(timeout time.Duration) *resty.Client {
	return resty.New().
		SetDebug(false).
		SetTimeout(timeout)
}
