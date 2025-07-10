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

func GZipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Проверяем, поддерживает ли клиент gzip сжатие для ответов
		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		// Проверяем, отправил ли клиент сжатые данные в запросе
		contentEncoding := req.Header.Get("Content-Encoding")
		isGzipRequest := strings.Contains(contentEncoding, "gzip")

		// Обрабатываем входящий сжатый запрос
		if isGzipRequest && req.Body != nil {
			gzipReader, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(w, "Invalid gzip data", http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			req.Body = gzipReader
			req.Header.Del("Content-Encoding")
		}

		// Обрабатываем исходящий ответ - сжимаем если клиент поддерживает
		if supportsGzip {
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			gzipRW := &gzipResponseWriter{
				ResponseWriter: w,
				gzipWriter:     gzipWriter,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(gzipRW, req)
		} else {
			next.ServeHTTP(w, req)
		}
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter  *gzip.Writer
	statusCode  int
	contentSize int
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	contentSize, err := w.gzipWriter.Write(data)
	w.contentSize += contentSize
	return contentSize, err
}

func (w *gzipResponseWriter) Close() error {
	return w.gzipWriter.Close()
}
