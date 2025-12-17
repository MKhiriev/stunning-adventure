package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
)

func (h *Handler) WithDecryption(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.WithDecryption").Msg("error reading http body")
			return
		}

		// pass through if empty body
		if len(body) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// decrypt message
		decryptedMessage, err := rsa.DecryptPKCS1v15(rand.Reader, h.privateKey, body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.WithDecryption").Msg("error decrypting http body")
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(decryptedMessage))

		next.ServeHTTP(w, r)
	})
}
