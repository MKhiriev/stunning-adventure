package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func GZip(next http.Handler) http.Handler {
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
