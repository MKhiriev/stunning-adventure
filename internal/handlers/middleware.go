package handlers

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}

	compressionResponseWriter struct {
		w  http.ResponseWriter
		zw *gzip.Writer
	}

	compressionsRequestReader struct {
		r  io.ReadCloser
		zr *gzip.Reader
	}
)

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

		h.Logger.Info().
			Str("uri", uri).
			Str("method", method).
			Int("status", responseData.status).
			Dur("duration", duration).
			Int("size", responseData.size).
			Msg("")
	}

	return http.HandlerFunc(logFn)
}

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

func WithGzipCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		contentTypesToCompress := []string{"application/json", "text/html", "text/html; charset=utf-8"}

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip || slices.Contains(contentTypesToCompress, r.Header.Get("Content-Type")) {
			cw := compressionWriter(w)
			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := compressionReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}

func compressionWriter(w http.ResponseWriter) *compressionResponseWriter {
	return &compressionResponseWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func compressionReader(r io.ReadCloser) (*compressionsRequestReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressionsRequestReader{
		r:  r,
		zr: zr,
	}, nil
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (c *compressionResponseWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressionResponseWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressionResponseWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressionResponseWriter) Close() error {
	return c.zw.Close()
}

func (c compressionsRequestReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressionsRequestReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
