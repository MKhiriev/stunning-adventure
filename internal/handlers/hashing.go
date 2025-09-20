package handlers

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/rs/zerolog"
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
		h.logger.Debug().Str("func", "*Handler.WithHashing").Str("hash from header", hashFromHeader).Str("hash from body", hex.EncodeToString(hashSliceFromBody)).Msg("")

		// compare hash from header with calculated hash from body
		if hashFromHeader != hex.EncodeToString(hashSliceFromBody) {
			h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("hash is not valid")
			http.Error(w, "hashes do not match", http.StatusBadRequest)
			return
		}

		h.logger.Debug().Str("func", "*Handler.WithHashing").Msg("metric hash is valid")

		// catching response from server
		rw := &responseWriterWithHash{ResponseWriter: w, hasher: h.hasher, logger: h.logger}

		// call another middleware or handler
		next.ServeHTTP(rw, req)
	})
}

// responseWriterWithHash catches response from server
type responseWriterWithHash struct {
	http.ResponseWriter
	hasher     *utils.Hasher
	logger     *zerolog.Logger
	statusCode int
}

func (rw *responseWriterWithHash) Write(p []byte) (int, error) {
	rw.logger.Debug().
		Str("func", "responseWriterWithHash.Write()").
		Bytes("writer bytes", p).Msg("")

	// hash body
	if responseHash, err := rw.hasher.HashByteSlice(p); err == nil {
		rw.Header().Set("HashSHA256", hex.EncodeToString(responseHash))
	}

	rw.WriteHeader(http.StatusOK)

	rw.logger.Debug().
		Str("func", "responseWriterWithHash.Write()").
		Any("header", rw.Header()).
		RawJSON("body", p).Msg("")

	return rw.ResponseWriter.Write(p)
}
