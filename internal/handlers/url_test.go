package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"urlshortener/internal/deps/mocks"
	"urlshortener/internal/storage/inmemory"

	"go.uber.org/mock/gomock"
)

//type testEnv struct {
//	mock...
//	ctrl...
//  HandlerURL
//}

func TestHandlerURL_SetURLwithJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockURLshortener(ctrl)

	tests := []struct {
		name         string
		setupMock    func()
		requestBody  string
		method       string // fixme: один и тот же, не меняется
		contentType  string
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful URL shortening",
			setupMock: func() {
				mockShortener.EXPECT().
					SetURL("https://example.com").
					Return("abc123", nil)
			},
			requestBody:  `{"url":"https://example.com"}`,
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusCreated,
			expectedBody: `{"result":"http://localhost:8080/abc123"}`,
		},
		{
			name:         "empty JSON body",
			setupMock:    func() {},
			requestBody:  "",
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid request body: EOF"}`,
		},
		{
			name:         "invalid JSON format",
			setupMock:    func() {},
			requestBody:  `{"url":`,
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid request body: unexpected EOF"}`,
		},
		{
			name:         "missing URL field",
			setupMock:    func() {},
			requestBody:  `{"invalid":"field"}`,
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"url is required"}`,
		},
		{
			name:         "empty URL",
			setupMock:    func() {},
			requestBody:  `{"url":""}`,
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"url is required"}`,
		},
		{
			name: "service error",
			setupMock: func() {
				mockShortener.EXPECT().
					SetURL("https://error.com").
					Return("", errors.New("service error"))
			},
			requestBody:  `{"url":"https://error.com"}`,
			method:       http.MethodPost,
			contentType:  "application/json",
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":"failed to shorten URL: service error"}`,
		},
		{
			name:         "wrong HTTP method",
			setupMock:    func() {},
			requestBody:  `{"url":"https://example.com"}`,
			method:       http.MethodGet,
			contentType:  "application/json",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: `{"error":"method not allowed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			h := &HandlerURL{
				service: mockShortener,
				baseURL: "localhost:8080",
			}

			req, err := http.NewRequest(tt.method, "/api/shorten", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			h.SetURLwithJSON(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			if tt.expectedBody != "" {
				// Сравниваем как строки, так как нам важно точное соответствие
				got := strings.TrimSpace(rr.Body.String())
				want := strings.TrimSpace(tt.expectedBody)
				if got != want {
					t.Errorf("handler returned unexpected body:\ngot  %v\nwant %v",
						got, want)
				}
			}
		})
	}
}

func TestHandlderURL_GetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockURLshortener(ctrl)

	tests := []struct {
		name         string
		setupMock    func()
		requestURL   string
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful redirect",
			setupMock: func() {
				mockShortener.EXPECT().
					GetURL("abc123").
					Return("https://example.com", nil)
			},
			requestURL:   "/abc123",
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name: "invalid token",
			setupMock: func() {
				mockShortener.EXPECT().
					GetURL("invalid").
					Return("", inmemory.GetErr)
			},
			requestURL:   "/invalid",
			expectedCode: http.StatusBadRequest,
			expectedBody: "GetURL Error(): " + inmemory.GetErr.Error(),
		},
		{
			name: "empty token",
			setupMock: func() {
				// Явно ожидаем вызов GetURL с пустой строкой
				mockShortener.EXPECT().
					GetURL("").
					Return("", errors.New("empty token"))
			},
			requestURL:   "/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "GetURL Error(): empty token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			h := &HandlerURL{
				service: mockShortener,
				baseURL: "localhost:8080",
			}

			req, err := http.NewRequest("GET", tt.requestURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			h.GetURL(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			if tt.expectedBody != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedBody) {
					t.Errorf("handler returned unexpected body: got %q want to contain %q",
						rr.Body.String(), tt.expectedBody)
				}
			}
		})
	}
}

func TestHandlderURL_SetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockURLshortener(ctrl)

	tests := []struct {
		name         string
		setupMock    func()
		requestBody  string
		expectedCode int
		expectedBody string
	}{
		{
			name: "successful URL shortening",
			setupMock: func() {
				mockShortener.EXPECT().
					SetURL("https://example.com").
					Return("abc123", nil)
			},
			requestBody:  "https://example.com",
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/abc123",
		},
		{
			name: "empty request body",
			setupMock: func() {
				// Мок не должен вызываться в этом случае
			},
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			expectedBody: "empty request body",
		},
		{
			name: "service error",
			setupMock: func() {
				mockShortener.EXPECT().
					SetURL("https://error.com").
					Return("", inmemory.SetErr)
			},
			requestBody:  "https://error.com",
			expectedCode: http.StatusBadRequest,
			expectedBody: "SetURL Error(): " + inmemory.SetErr.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// {
			// Настраиваем мок
			tt.setupMock()

			// Создаем обработчик с моком

			h := &HandlerURL{
				service: mockShortener,
				baseURL: "localhost:8080",
			}

			// } => testEnv

			// Создаем запрос с телом
			req, err := http.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			h.SetURL(rr, req)

			// Проверяем статус код
			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			// Проверяем тело ответа
			if tt.expectedBody != "" {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v",
						rr.Body.String(), tt.expectedBody)
				}
			}
		})
	}
}

func TestHandlderURL_DefaultURL(t *testing.T) {
	// Не нужны моки, так как метод не использует сервис
	h := &HandlerURL{
		service: nil,
		baseURL: "localhost:8080",
	}

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "GET request",
			method:       "GET",
			path:         "/unknown",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad Request (400)\nMethod: GET\nPath: /unknown",
		},
		{
			name:         "POST request",
			method:       "POST",
			path:         "/another",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad Request (400)\nMethod: POST\nPath: /another",
		},
		{
			name:         "PUT request",
			method:       "PUT",
			path:         "/test",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Bad Request (400)\nMethod: PUT\nPath: /test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			h.DefaultURL(rr, req)

			// Проверяем статус код
			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			// Проверяем тело ответа
			if got := rr.Body.String(); got != tt.expectedBody {
				t.Errorf("handler returned unexpected body:\ngot:\n%v\nwant:\n%v",
					got, tt.expectedBody)
			}

			// Дополнительно проверяем Content-Type
			if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
				t.Errorf("handler returned wrong content type: got %v want text/plain",
					contentType)
			}
		})
	}
}
