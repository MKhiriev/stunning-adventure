package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ========== ParseTrustedSubnet Tests ==========

func TestParseTrustedSubnet_Success_IPv4(t *testing.T) {
	ipNet, err := ParseTrustedSubnet("192.168.1.0/24")

	require.NoError(t, err)
	require.NotNil(t, ipNet)
	assert.Equal(t, "192.168.1.0/24", ipNet.String())
}

func TestParseTrustedSubnet_Success_IPv6(t *testing.T) {
	ipNet, err := ParseTrustedSubnet("2001:db8::/32")

	require.NoError(t, err)
	require.NotNil(t, ipNet)
	assert.Equal(t, "2001:db8::/32", ipNet.String())
}

func TestParseTrustedSubnet_Success_SingleIPv4(t *testing.T) {
	ipNet, err := ParseTrustedSubnet("192.168.1.100/32")

	require.NoError(t, err)
	require.NotNil(t, ipNet)
	assert.Equal(t, "192.168.1.100/32", ipNet.String())
}

func TestParseTrustedSubnet_Success_SingleIPv6(t *testing.T) {
	ipNet, err := ParseTrustedSubnet("2001:db8::1/128")

	require.NoError(t, err)
	require.NotNil(t, ipNet)
	assert.Equal(t, "2001:db8::1/128", ipNet.String())
}

func TestParseTrustedSubnet_EmptyString_ReturnsNil(t *testing.T) {
	ipNet, err := ParseTrustedSubnet("")

	require.NoError(t, err)
	assert.Nil(t, ipNet, "empty CIDR should return nil without error")
}

func TestParseTrustedSubnet_Error_InvalidCIDR(t *testing.T) {
	testCases := []struct {
		name string
		cidr string
	}{
		{"invalid format", "192.168.1.0"},
		{"invalid IP", "999.999.999.999/24"},
		{"invalid mask", "192.168.1.0/33"},
		{"garbage", "not-a-cidr"},
		{"missing mask", "192.168.1.0/"},
		{"negative mask", "192.168.1.0/-1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipNet, err := ParseTrustedSubnet(tc.cidr)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid trusted subnet CIDR")
			assert.Nil(t, ipNet)
		})
	}
}

func TestParseTrustedSubnet_LargeSubnets(t *testing.T) {
	t.Run("IPv4 /0", func(t *testing.T) {
		ipNet, err := ParseTrustedSubnet("0.0.0.0/0")
		require.NoError(t, err)
		require.NotNil(t, ipNet)
		assert.Equal(t, "0.0.0.0/0", ipNet.String())
	})

	t.Run("IPv6 /0", func(t *testing.T) {
		ipNet, err := ParseTrustedSubnet("::/0")
		require.NoError(t, err)
		require.NotNil(t, ipNet)
		assert.Equal(t, "::/0", ipNet.String())
	})
}

// ========== TrustedSubnetInterceptor Tests ==========

func TestTrustedSubnetInterceptor_Success_ValidIPv4(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "192.168.1.100",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerCalled, "handler should be called")
}

func TestTrustedSubnetInterceptor_Success_ValidIPv6(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("2001:db8::/32")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "2001:db8::1",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_Success_NilTrustedSubnet(t *testing.T) {
	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(nil, logger)

	// No metadata needed when trustedSubnet is nil
	ctx := context.Background()

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerCalled, "handler should be called when trustedSubnet is nil")
}

func TestTrustedSubnetInterceptor_Success_EdgeOfSubnet(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	testCases := []string{
		"192.168.1.0",   // first IP
		"192.168.1.1",   // second IP
		"192.168.1.254", // second to last
		"192.168.1.255", // last IP (broadcast)
	}

	for _, ip := range testCases {
		t.Run(ip, func(t *testing.T) {
			md := metadata.New(map[string]string{
				"X-Real-IP": ip,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			require.NoError(t, err)
			assert.Equal(t, "success", resp)
			assert.True(t, handlerCalled)
		})
	}
}

func TestTrustedSubnetInterceptor_Error_MissingMetadata(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	ctx := context.Background() // No metadata

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.Error(t, err)
	assert.ErrorIs(t, err, errMissingMetadata)
	assert.Nil(t, resp)
	assert.False(t, handlerCalled, "handler should not be called")
}

func TestTrustedSubnetInterceptor_Error_MissingXRealIP(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"Other-Header": "value",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.Error(t, err)
	assert.ErrorIs(t, err, errMissingRealIPMetadata)
	assert.Nil(t, resp)
	assert.False(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_Error_EmptyXRealIP(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.Error(t, err)
	assert.ErrorIs(t, err, errEmptyRealIPMetadata)
	assert.Nil(t, resp)
	assert.False(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_Error_InvalidIPAddress(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	invalidIPs := []string{
		"not-an-ip",
		"999.999.999.999",
		"192.168.1",
		"192.168.1.1.1",
		"::gggg",
	}

	for _, invalidIP := range invalidIPs {
		t.Run(invalidIP, func(t *testing.T) {
			md := metadata.New(map[string]string{
				"X-Real-IP": invalidIP,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			require.Error(t, err)
			assert.ErrorIs(t, err, errInvalidIPAddress)
			assert.Nil(t, resp)
			assert.False(t, handlerCalled)
		})
	}
}

func TestTrustedSubnetInterceptor_Error_IPNotInTrustedSubnet(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	untrustedIPs := []string{
		"192.168.2.1",   // different subnet
		"10.0.0.1",      // completely different
		"192.168.0.255", // just before subnet
		"192.168.2.0",   // just after subnet
		"8.8.8.8",       // public IP
	}

	for _, untrustedIP := range untrustedIPs {
		t.Run(untrustedIP, func(t *testing.T) {
			md := metadata.New(map[string]string{
				"X-Real-IP": untrustedIP,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			require.Error(t, err)
			assert.ErrorIs(t, err, errNotFromTrustedSubnet)
			assert.Nil(t, resp)
			assert.False(t, handlerCalled)
		})
	}
}

func TestTrustedSubnetInterceptor_IPv6_NotInTrustedSubnet(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("2001:db8::/32")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "2001:db9::1", // different subnet
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.Error(t, err)
	assert.ErrorIs(t, err, errNotFromTrustedSubnet)
	assert.Nil(t, resp)
	assert.False(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_MultipleXRealIPValues(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	// gRPC metadata can have multiple values for the same key
	md := metadata.MD{
		"x-real-ip": []string{"192.168.1.100", "192.168.1.101"},
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	// Should use the first value
	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_HandlerReturnsError(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "192.168.1.100",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerError := assert.AnError
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, handlerError
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.Error(t, err)
	assert.Equal(t, handlerError, err)
	assert.Nil(t, resp)
}

func TestTrustedSubnetInterceptor_Localhost(t *testing.T) {
	t.Run("IPv4 localhost in 127.0.0.0/8", func(t *testing.T) {
		_, trustedSubnet, err := net.ParseCIDR("127.0.0.0/8")
		require.NoError(t, err)

		logger := newTestLogger()
		interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

		md := metadata.New(map[string]string{
			"X-Real-IP": "127.0.0.1",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		handlerCalled := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return "success", nil
		}

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

		require.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, handlerCalled)
	})

	t.Run("IPv6 localhost", func(t *testing.T) {
		_, trustedSubnet, err := net.ParseCIDR("::1/128")
		require.NoError(t, err)

		logger := newTestLogger()
		interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

		md := metadata.New(map[string]string{
			"X-Real-IP": "::1",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		handlerCalled := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return "success", nil
		}

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

		require.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, handlerCalled)
	})
}

func TestTrustedSubnetInterceptor_CaseInsensitiveHeader(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	// gRPC metadata keys are case-insensitive
	md := metadata.MD{
		"x-real-ip": []string{"192.168.1.100"}, // lowercase
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerCalled)
}

func TestTrustedSubnetInterceptor_LargeSubnet(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("0.0.0.0/0")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	// Any IP should be allowed
	testIPs := []string{
		"1.2.3.4",
		"192.168.1.1",
		"8.8.8.8",
		"255.255.255.255",
	}

	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			md := metadata.New(map[string]string{
				"X-Real-IP": ip,
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handlerCalled := false
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			require.NoError(t, err)
			assert.Equal(t, "success", resp)
			assert.True(t, handlerCalled)
		})
	}
}

func TestTrustedSubnetInterceptor_ConcurrentRequests(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			md := metadata.New(map[string]string{
				"X-Real-IP": "192.168.1.100",
			})
			ctx := metadata.NewIncomingContext(context.Background(), md)

			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

			require.NoError(t, err)
			assert.Equal(t, "success", resp)

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestTrustedSubnetInterceptor_ContextPropagation(t *testing.T) {
	_, trustedSubnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	logger := newTestLogger()
	interceptor := TrustedSubnetInterceptor(trustedSubnet, logger)

	md := metadata.New(map[string]string{
		"X-Real-IP": "192.168.1.100",
	})
	type testKey string
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx = context.WithValue(ctx, testKey("test_key"), "test_value")

	handlerReceivedContext := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify context is propagated
		if ctx.Value(testKey("test_key")) == "test_value" {
			handlerReceivedContext = true
		}
		return "success", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	assert.True(t, handlerReceivedContext, "context should be propagated to handler")
}
