package handlers

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"
)

func (h *Handler) WithHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// if we don't have hashkey to check hash or do not have an http-body - do nothing
		if h.hasher == nil || req.Body == nil {
			next.ServeHTTP(w, req)
			return
		}
		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("validating metric hash")

		// check if hash exists in header
		hashFromHeader := req.Header.Get("HashSHA256")
		h.logger.Debug().Str("func", "*Handler.WithHashing").Str("hash from header", hashFromHeader).Msg("")

		if hashFromHeader == "" {
			http.Error(w, "empty hash", http.StatusBadRequest)
			return
		}

		// read request body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		h.logger.Debug().Str("func", "*Handler.WithHashing").Bytes("http body", body).Msg("")

		// return body to the request
		req.Body = io.NopCloser(bytes.NewBuffer(body))

		// get hash body
		hashSliceFromBody, err := h.hasher.HashByteSlice(body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		h.logger.Debug().Str("func", "*Handler.WithHashing").Str("hash from header", hashFromHeader).Bytes("hash from body", hashSliceFromBody).Msg("")

		// compare hash from header with calculated hash from body
		if hashFromHeader != hex.EncodeToString(hashSliceFromBody) {
			http.Error(w, "hashes do not match", http.StatusBadRequest)
			return
		}

		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("metric hash is valid")

		next.ServeHTTP(w, req)
	})
}
