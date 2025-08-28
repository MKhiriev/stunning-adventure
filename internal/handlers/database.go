package handlers

import (
	"context"
	"net/http"
)

func (h *Handler) DatabaseConnectionCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h.dbPingService.Ping(context.Background()); err != nil {
			h.logger.Error().Str("middleware", "DatabaseConnectionCheck").Msg("database is not connected")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}
