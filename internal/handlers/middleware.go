package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/context"
)

const timeout = 10 * time.Second

func (h *Handler) WithLogging(handler http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		handler.ServeHTTP(&lw, r)

		duration := time.Since(start)

		h.logger.Info().
			Str("uri", uri).
			Str("method", method).
			Int("status", responseData.status).
			Dur("duration", duration).
			Int("size", responseData.size).
			Msg("")
	}

	return http.HandlerFunc(logFn)
}

// CheckHTTPMethod метод для возвращения статус кода 404 в случае обращения к зарегистрированному роуту с неправильным методом
func CheckHTTPMethod(router *chi.Mux) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedURL := r.URL.Path
		requestedHTTPMethod := r.Method

		allRoutes := router.Routes()
		var foundRoute chi.Route
		for _, route := range allRoutes {
			if route.Pattern == requestedURL {
				foundRoute = route
				break
			}
		}

		if _, ok := foundRoute.Handlers[requestedHTTPMethod]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		router.ServeHTTP(w, r)
	}
}

func WithContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFunc := context.WithTimeout(r.Context(), timeout)
		defer cancelFunc()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
