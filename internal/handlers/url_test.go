package handlers

import (
	"net/http"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				code:        200,
				response:    `{"status":"ok"}`,
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// здесь будет запрос и проверка ответа
		})
	}
}

func TestHandlderURL_GetURL(t *testing.T) {
	type fields struct {
		service URLshortener
		BaseURL string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HandlderURL{
				service: tt.fields.service,
				BaseURL: tt.fields.BaseURL,
			}
			h.GetURL(tt.args.w, tt.args.r)
		})
	}
}

func TestHandlderURL_SetURL(t *testing.T) {
	type fields struct {
		service URLshortener
		BaseURL string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HandlderURL{
				service: tt.fields.service,
				BaseURL: tt.fields.BaseURL,
			}
			h.SetURL(tt.args.w, tt.args.r)
		})
	}
}
