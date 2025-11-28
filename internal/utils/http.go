package utils

import "github.com/go-resty/resty/v2"

// NewHTTPClient constructs and returns a Resty HTTP client with debugging disabled.
func NewHTTPClient() *resty.Client {
	return resty.New().SetDebug(false)
}
