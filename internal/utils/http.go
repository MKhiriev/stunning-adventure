package utils

import (
	"fmt"
	"net"
	"time"

	"github.com/go-resty/resty/v2"
)

// NewHTTPClient constructs and returns a Resty HTTP client with debugging disabled.
func NewHTTPClient(timeout time.Duration) *resty.Client {
	return resty.New().
		SetDebug(false).
		SetTimeout(timeout)
}

func GetLocalIP(address string) (net.IP, error) {
	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, fmt.Errorf("error dialing server: %w", err)
	}
	defer conn.Close()

	localAddress, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddress == nil {
		return nil, fmt.Errorf("unable to resolve: %w", err)
	}

	return localAddress.IP, nil
}
