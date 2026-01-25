package grpc

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TrustedSubnetInterceptor(trustedSubnet *net.IPNet, logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error().Msg("failed to get metadata from context")
			return nil, errMissingMetadata
		}

		realIPValues := md.Get("x-real-ip")
		if len(realIPValues) == 0 {
			logger.Error().Msg("x-real-ip header is missing")
			return nil, errMissingRealIpMetadata
		}

		realIP := realIPValues[0]
		if realIP == "" {
			logger.Error().Msg("x-real-ip header is empty")
			return nil, errEmptyRealIpMetadata
		}

		ip, err := netip.ParseAddr(realIP)
		if err != nil {
			logger.Err(err).Str("ip", realIP).Msg("invalid IP address in x-real-ip")
			return nil, errInvalidIPAddress
		}

		if !trustedSubnet.Contains(ip.AsSlice()) {
			logger.Error().
				Str("ip", realIP).
				Str("trusted_subnet", trustedSubnet.String()).
				Msg("IP address is not from trusted subnet")
			return nil, errNotFromTrustedSubnet
		}

		logger.Debug().Str("ip", realIP).Msg("IP address verified in trusted subnet")
		return handler(ctx, req)
	}
}

func ParseTrustedSubnet(cidr string) (*net.IPNet, error) {
	if cidr == "" {
		return nil, nil
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid trusted subnet CIDR: %w", err)
	}

	return ipNet, nil
}
