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
	HeaderLocation        = "Location"

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

// WriteTextResponse writes a plain text response
func WriteTextResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func WriteBadRequestError(w http.ResponseWriter, details string) {
	response := fmt.Sprintf("Bad Request (%d)\n%s", http.StatusBadRequest, details)
	WriteTextError(w, http.StatusBadRequest, response)
}

// WriteTextError writes a plain text error response
func WriteTextError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", MIMETextPlain)
	w.WriteHeader(status)
	w.Write([]byte(message))
}

// WriteJSONError writes a JSON error response
func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: message})
}

// WriteJSONResponse writes a JSON response
func WriteJSONResponse[T any](w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", MIMEApplicationJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteRedirect writes a redirect response
func WriteRedirect(w http.ResponseWriter, location string, permanent bool) {
	status := http.StatusTemporaryRedirect
	if permanent {
		status = http.StatusPermanentRedirect
	}

	w.Header().Set(HeaderLocation, location)
	w.WriteHeader(status)
}

// BuildShortURL constructs a short URL from base and ID
func BuildShortURL(urlroot, id string) string {
	return fmt.Sprintf("http://%s/%s", urlroot, id)
}
