package handlers

import (
	"io"
	"net/http"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

func (h *Handler) WithDecryption(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Msg("error reading http body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// pass through if empty body
		if len(body) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// decrypt message
		decryptedMessage, err := utils.DecryptData(body, h.privateKey)
		if err != nil {
			h.logger.Err(err).Msg("error decrypting http body")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// set a decrypted data to the body
		r.Body = io.NopCloser(decryptedMessage)

		next.ServeHTTP(w, r)
	})
}
