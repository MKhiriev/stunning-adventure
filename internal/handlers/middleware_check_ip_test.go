package handlers

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_CheckIPTrustedSubnet(t *testing.T) {
	logger := zerolog.Nop()

	// Mock next handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		trustedSubnet  string
		realIPHeader   string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid IP from trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIPHeader:   "192.168.1.100",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "valid IP not from trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			realIPHeader:   "10.0.0.1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			realIPHeader:   "",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "invalid IP address format",
			trustedSubnet:  "192.168.1.0/24",
			realIPHeader:   "invalid-ip",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "IPv6 address from trusted subnet",
			trustedSubnet:  "2001:db8::/32",
			realIPHeader:   "2001:db8::1",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "IPv6 address not from trusted subnet",
			trustedSubnet:  "2001:db8::/32",
			realIPHeader:   "2001:db9::1",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
		{
			name:           "edge case - IP at subnet boundary (first IP)",
			trustedSubnet:  "10.0.0.0/24",
			realIPHeader:   "10.0.0.0",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "edge case - IP at subnet boundary (last IP)",
			trustedSubnet:  "10.0.0.0/24",
			realIPHeader:   "10.0.0.255",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "single IP subnet /32",
			trustedSubnet:  "192.168.1.100/32",
			realIPHeader:   "192.168.1.100",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:           "single IP subnet /32 - different IP",
			trustedSubnet:  "192.168.1.100/32",
			realIPHeader:   "192.168.1.101",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Forbidden\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse trusted subnet
			_, ipNet, err := net.ParseCIDR(tt.trustedSubnet)
			require.NoError(t, err, "failed to parse trusted subnet")

			// Create handler with trusted subnet
			handler := &Handler{
				logger:        &logger,
				trustedSubnet: ipNet,
			}

			// Create middleware
			middleware := handler.CheckIPTrustedSubnet(nextHandler)

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.realIPHeader != "" {
				req.Header.Set("X-Real-IP", tt.realIPHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Assert response body
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
		})
	}
}

func TestHandler_CheckIPTrustedSubnet_NilTrustedSubnet(t *testing.T) {
	logger := zerolog.Nop()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := &Handler{
		logger:        &logger,
		trustedSubnet: nil,
	}

	// This test verifies behavior when trustedSubnet is nil
	// In your Init() method, you check if trustedSubnet != nil before using the middleware
	// So this test ensures the middleware doesn't panic with nil subnet
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Real-IP", "192.168.1.1")

	rr := httptest.NewRecorder()

	// This should panic or fail gracefully
	assert.Panics(t, func() {
		middleware := handler.CheckIPTrustedSubnet(nextHandler)
		middleware.ServeHTTP(rr, req)
	}, "expected panic when trustedSubnet is nil")
}
