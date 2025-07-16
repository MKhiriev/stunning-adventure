package handlers

import "net/http"

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	// TODO decide how to fix redundant nil check
	//if h.db == nil {
	//	h.logger.Error().Str("func", "*Handler.Ping").Msg("db connection is nil")
	//	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	//	return
	//}
	err := h.db.Ping(r.Context())
	if err != nil {
		h.logger.Err(err).Str("func", "*Handler.Ping").Msg("database ping error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.logger.Info().Str("func", "*Handler.Ping").Msg("database pinged successfully")
	w.WriteHeader(http.StatusOK)
}
