package handlers

import "net/http"

// URLGetter обрабатывает запросы на получение оригинального URL по короткому идентификатору
type URLGetter interface {
	GetURL(w http.ResponseWriter, r *http.Request)
}

// URLTextSetter обрабатывает запросы на создание короткого URL с текстовым содержимым
type URLTextSetter interface {
	SetURLFromText(w http.ResponseWriter, r *http.Request)
}

// URLJSONSetter обрабатывает запросы на создание короткого URL с JSON содержимым
type URLJSONSetter interface {
	SetURLFromJSON(w http.ResponseWriter, r *http.Request)
}

// DefaultHandler обрабатывает все остальные запросы
type DefaultHandler interface {
	HandleDefault(w http.ResponseWriter, r *http.Request)
}
