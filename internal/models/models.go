package models

type SetURLJSONRequest struct {
	URL string `json:"url"`
}

type SetURLJSONResponse struct {
	URLShort string `json:"urlshort"`
}

type SetURLJSONErrorResponse struct {
	Error string `json:"error"`
}
