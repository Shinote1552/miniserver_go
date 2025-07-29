package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"urlshortener/internal/httputils"
)

func MiddlewareCompressing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Распаковка запроса
			if err := decompressRequest(r); err != nil {
				http.Error(w, "invalid gzip data", http.StatusBadRequest)
				return
			}

			// Проверка необходимости сжатия ответа
			if !acceptsGzip(r) || !isCompressible(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Подготовка сжатого ответа
			gz := gzip.NewWriter(w)
			defer gz.Close()

			w.Header().Set(httputils.HeaderContentEncoding, httputils.EncodingGzip)
			w.Header().Del(httputils.HeaderContentLength)

			next.ServeHTTP(&gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}, r)
		})
	}
}

func decompressRequest(r *http.Request) error {
	if !strings.Contains(r.Header.Get(httputils.HeaderContentEncoding), httputils.EncodingGzip) {
		return nil
	}

	gz, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	r.Body = gz
	return nil
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get(httputils.HeaderAcceptEncoding), httputils.EncodingGzip)
}

func isCompressible(r *http.Request) bool {
	contentType := r.Header.Get(httputils.HeaderContentType)
	return strings.HasPrefix(contentType, httputils.MIMEApplicationJSON) ||
		strings.HasPrefix(contentType, httputils.MIMETextHTML) ||
		strings.HasPrefix(contentType, httputils.MIMETextPlain)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
