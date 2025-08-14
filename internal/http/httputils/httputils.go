package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// MIME: https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/MIME_types/Common_types

const (
	HeaderContentType     = "Content-Type"
	HeaderContentEncoding = "Content-Encoding"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentLength   = "Content-Length"
	HeaderUserAgent       = "User-Agent"

	MIMEApplicationJSON       = "application/json"
	MIMETextHTML              = "text/html"
	MIMETextPlain             = "text/plain"
	MIMEGZipCompressedArchive = "application/gzip"

	EncodingGzip = "gzip"
)

var (
	ErrInvalidData = errors.New("invalid data")
	ErrUnfound     = errors.New("unfound data")
	ErrEmpty       = errors.New("storage is empty")
	ErrConflict    = errors.New("url already exists with different value")
)

func WriteTextError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}

func WriteJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func BuildShortURL(urlroot, id string) string {
	return fmt.Sprintf("http://%s/%s", urlroot, id)
}
