package handlers

import (
	"net/http"
	"net/netip"
)

const realIPHeader = "X-Real-IP"

func (h *Handler) CheckIPTrustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realIP := r.Header.Get(realIPHeader)
		if realIP == "" {
			h.logger.Error().Msgf("header `%s`: ip address is empty", realIPHeader)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		ip, err := netip.ParseAddr(realIP)
		if err != nil {
			h.logger.Err(err).Msgf("header `%s`: invalid ip address", realIPHeader)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		fromTrustedNetwork := h.trustedSubnet.Contains(ip.AsSlice())
		if !fromTrustedNetwork {
			h.logger.Err(err).Msgf("header `%s`: ip address is not from trusted network", realIPHeader)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
