package handlers

import (
	"net/http"
	"time"
)

func (h *Handler) WithLogging(handler http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		handler.ServeHTTP(lw, r)

		duration := time.Since(start)

		h.logger.Info().
			Str("uri", uri).
			Str("method", method).
			Int("status", lw.status).
			Dur("duration", duration).
			Int("size", lw.size).
			Send()
	}

	return http.HandlerFunc(logFn)
}

type responseData struct {
	status int
	size   int
	body   []byte
}

type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	status              int
	size                int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
