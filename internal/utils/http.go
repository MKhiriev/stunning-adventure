package utils

import "github.com/go-resty/resty/v2"

func NewHTTPClient() *resty.Client {
	return resty.New().SetDebug(false)
}
