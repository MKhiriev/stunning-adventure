package utils

import (
	"fmt"
	"net"
	"strings"
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
	conn, err := net.Dial("udp4", address)
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

func CheckIfValidIPAddress(address string) error {
	ip := strings.Split(address, ":")[0]

	_, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		return fmt.Errorf("incorrect ip: %v\n", err)
	}

	return nil
}
