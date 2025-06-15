package handlers

import "urlshortener/internal/service"

type HandlderURL struct {
	service *service.URLshortener
}

func NewHandlderURL(service *service.URLshortener) *HandlderURL {
	return &HandlderURL{
		service: service,
	}
}

// GET
func (*HandlderURL) GetURL() {

}

// POST
func (*HandlderURL) SetURL() {

}
