package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"urlshortener/internal/handlers/mocks"
	"urlshortener/internal/storage/inmemory"

	"go.uber.org/mock/gomock"
)

func TestHandlderURL_GetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockShortener := mocks.NewMockURLshortener(ctrl)

	tests := []struct {
		name         string
		setupMock    func()
		requestURL   string
		expectedCode int
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.setupMock()

			// Создаем обработчик с моком
			h := &HandlderURL{
				service: mockShortener,
				BaseURL: "localhost:8080",
			}

			// Создаем запрос
			req, err := http.NewRequest("GET", tt.requestURL, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			h.GetURL(rr, req)

			// Проверяем статус код
			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedCode)
			}

			// Для редиректа проверяем заголовок Location
			if tt.expectedCode == http.StatusTemporaryRedirect {
				location := rr.Header().Get("Location")
				if location != "https://example.com" {
					t.Errorf("handler returned wrong location header: got %v want %v",
						location, "https://example.com")
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
			// Настраиваем мок
			tt.setupMock()

			// Создаем обработчик с моком
			h := &HandlderURL{
				service: mockShortener,
				BaseURL: "localhost:8080",
			}

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
