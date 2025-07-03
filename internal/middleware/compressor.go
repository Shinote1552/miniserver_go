package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"urlshortener/internal/httputils"
)

// MiddlewareCompressing возвращает middleware для gzip сжатия/распаковки
func MiddlewareCompressing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Обработка входящего сжатого контента
			if err := decompressRequest(r); err != nil {
				http.Error(w, "invalid gzip data", http.StatusBadRequest)
				return
			}

			// Обработка сжатия ответа
			if acceptsGzip(r) && isCompressible(r) {
				compressResponse(w, next)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// decompressRequest распаковывает входящий gzip-контент
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

// acceptsGzip проверяет поддержку gzip клиентом
func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get(httputils.HeaderAcceptEncoding), httputils.EncodingGzip)
}

// isCompressible проверяет нужно ли сжимать ответ
func isCompressible(r *http.Request) bool {
	contentType := r.Header.Get(httputils.HeaderContentType)
	return strings.HasPrefix(contentType, httputils.MIMEApplicationJSON) ||
		strings.HasPrefix(contentType, httputils.MIMETextHTML) ||
		strings.HasPrefix(contentType, httputils.MIMETextPlain)
}

// compressResponse сжимает ответ
func compressResponse(w http.ResponseWriter, next http.Handler) {
	gz := gzip.NewWriter(w)
	defer gz.Close()

	w.Header().Set(httputils.HeaderContentEncoding, httputils.EncodingGzip)
	w.Header().Del(httputils.HeaderContentLength)

	next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, nil)
}

// gzipResponseWriter минимальная обёртка для сжатия ответа
type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
