package handlers

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
)

func (h *Handler) WithHashing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.hashKey == "" {
			h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("no hashing will be done")
			next.ServeHTTP(w, r)
			return
		}

		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("WithHashing called")
		hashFromHeader := r.Header.Get("HashSHA256")
		if hashFromHeader == "" {
			next.ServeHTTP(w, r)
			return
		}
		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("checking hash begins")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Err(err).Str("func", "*Handler.WithHashing").Msg("failed to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		hashedBody := hex.EncodeToString(utils.Hash(body, h.hashKey))
		if hashedBody != hashFromHeader {
			h.logger.Error().Str("func", "*Handler.WithHashing").
				Str("hash from header", hashFromHeader).
				Str("hashed body", hashedBody).
				Msg("hashes are not equal")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.logger.Debug().Str("func", "*Handler.WithHashing").
			Str("hash from header", hashFromHeader).
			Str("hashed body", hashedBody).
			Msg("hashes are equal")
		rw := &HashingResponseWriter{
			ResponseWriter: w,
			hashKey:        h.hashKey,
		}

		next.ServeHTTP(rw, r)

		h.logger.Debug().
			Bytes("body", rw.responseData.body).
			Int("status code", rw.responseData.status).
			Any("headers", w.Header()).
			Msg("preparing to write response")
	})
}

type HashingResponseWriter struct {
	http.ResponseWriter
	responseData
	hashKey string
}

func (w *HashingResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
}

func (w *HashingResponseWriter) Write(data []byte) (int, error) {
	w.responseData.body = data
	hashFromResponseBody := hex.EncodeToString(utils.Hash(w.responseData.body, w.hashKey))
	w.Header().Set("HashSHA256", hashFromResponseBody)
	w.ResponseWriter.WriteHeader(w.responseData.status)
	size, err := w.ResponseWriter.Write(data)
	w.responseData.size += size
	return size, err
}
