package middleware

import "net/http"

type ResponseRecorder struct {
	ResponseWriter http.ResponseWriter
	StatusCode     int
	Size           int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
	}
}

// WriteHeader перехватывает и сохраняет статус код
func (r *ResponseRecorder) WriteHeader(status int) {
	r.StatusCode = status
	r.ResponseWriter.WriteHeader(status)
}

// Write перехватывает и сохраняет размер ответа
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	return size, err
}

// Header реализует интерфейс http.ResponseWriter
func (r *ResponseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}
