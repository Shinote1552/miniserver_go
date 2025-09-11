package url_shortener

import (
	"context"
	"fmt"
	"testing"
	"time"
	"urlshortener/internal/domain/models"
	"urlshortener/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestURLShortener_GetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockURLStorage(ctrl)
	service := NewServiceURLShortener(mockStorage, "http://short")

	tests := []struct {
		name        string
		shortKey    string
		mockSetup   func()
		wantURL     models.ShortenedLink
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "Успешное получение URL",
			shortKey: "abc123",
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), "abc123").
					Return(models.ShortenedLink{
						OriginalURL: "http://long.url",
						ShortCode:   "abc123",
					}, nil)
			},
			wantURL: models.ShortenedLink{
				OriginalURL: "http://long.url",
				ShortCode:   "abc123",
			},
		},
		{
			name:     "Пустой shortKey",
			shortKey: "",
			mockSetup: func() {
				// Нет вызовов к хранилищу
			},
			wantErr:     true,
			expectedErr: models.ErrInvalidData,
		},
		{
			name:     "URL не найден",
			shortKey: "notfound",
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), "notfound").
					Return(models.ShortenedLink{}, models.ErrUnfound)
			},
			wantErr:     true,
			expectedErr: models.ErrUnfound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := service.GetURL(context.Background(), tt.shortKey)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got)
		})
	}
}

func TestURLShortener_GetShortURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		shortKey string
		want     string
	}{
		{
			name:     "Формирование короткого URL",
			baseURL:  "http://short",
			shortKey: "abc123",
			want:     "http://short/abc123",
		},
		{
			name:     "Пустой shortKey",
			baseURL:  "http://short",
			shortKey: "",
			want:     "http://short/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewServiceURLShortener(nil, tt.baseURL)
			got := service.GetShortURL(tt.shortKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestURLShortener_SetURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockURLStorage(ctrl)
	service := NewServiceURLShortener(mockStorage, "http://short")

	tests := []struct {
		name        string
		input       models.ShortenedLink
		mockSetup   func()
		wantURL     models.ShortenedLink
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Успешное создание короткой ссылки",
			input: models.ShortenedLink{
				OriginalURL: "http://long.url",
				UserID:      1,
			},
			mockSetup: func() {
				// Генерация токена - проверка уникальности (может вызываться несколько раз)
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), gomock.Any()).
					Return(models.ShortenedLink{}, models.ErrUnfound).AnyTimes()

				// Создание записи - возвращаем созданный URL
				mockStorage.EXPECT().
					ShortenedLinkCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, url models.ShortenedLink) (models.ShortenedLink, error) {
						return models.ShortenedLink{
							OriginalURL: url.OriginalURL,
							ShortCode:   url.ShortCode, // сокращенная ссылка будет сгенерирован в generateUniqueToken
							UserID:      url.UserID,
							CreatedAt:   url.CreatedAt,
						}, nil
					})
			},
			wantURL: models.ShortenedLink{
				OriginalURL: "http://long.url",
				UserID:      1,
			},
			wantErr: false,
		},
		{
			name: "URL уже существует",
			input: models.ShortenedLink{
				OriginalURL: "http://existing.url",
				UserID:      1,
			},
			mockSetup: func() {
				// Генерация токена - проверка уникальности (может вызываться несколько раз)
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), gomock.Any()).
					Return(models.ShortenedLink{}, models.ErrUnfound).AnyTimes()

				// Попытка создания записи - возвращаем конфликт и существующий URL
				mockStorage.EXPECT().
					ShortenedLinkCreate(gomock.Any(), gomock.Any()).
					Return(models.ShortenedLink{
						ID:          1,
						OriginalURL: "http://existing.url",
						ShortCode:   "existing",
						UserID:      1,
						CreatedAt:   time.Now(),
					}, models.ErrConflict) // Важно: вернуть ошибку!
			},
			wantErr:     true,
			expectedErr: models.ErrConflict,
		},
		{
			name: "Невалидные данные - пустой URL",
			input: models.ShortenedLink{
				OriginalURL: "",
				UserID:      1,
			},
			mockSetup: func() {
				// Нет вызовов к хранилищу для невалидных данных
			},
			wantErr:     true,
			expectedErr: models.ErrInvalidData,
		},
		{
			name: "Невалидные данные - нулевой UserID",
			input: models.ShortenedLink{
				OriginalURL: "http://valid.url",
				UserID:      0,
			},
			mockSetup: func() {
				// Нет вызовов к хранилищу для невалидных данных
			},
			wantErr:     true,
			expectedErr: models.ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := service.SetURL(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				// Для случая конфликта проверяем, что возвращается существующий URL
				if tt.expectedErr == models.ErrConflict {
					assert.NotEmpty(t, got.ShortCode)
					assert.Equal(t, tt.input.OriginalURL, got.OriginalURL)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input.OriginalURL, got.OriginalURL)
			assert.Equal(t, tt.input.UserID, got.UserID)
			assert.NotEmpty(t, got.ShortCode)
			assert.False(t, got.CreatedAt.IsZero())
		})
	}
}

func TestURLShortener_BatchCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockURLStorage(ctrl)
	service := NewServiceURLShortener(mockStorage, "http://short")

	tests := []struct {
		name        string
		input       []models.ShortenedLink
		mockSetup   func()
		wantResult  []models.ShortenedLink
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Успешное пакетное создание",
			input: []models.ShortenedLink{
				{OriginalURL: "http://url1", UserID: 1},
				{OriginalURL: "http://url2", UserID: 1},
			},
			mockSetup: func() {
				// Проверка существующих URL
				mockStorage.EXPECT().
					ShortenedLinkBatchExists(gomock.Any(), []string{"http://url1", "http://url2"}).
					Return([]models.ShortenedLink{}, nil)

				// Проверка существующих shortCode после Генерации shortCode
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), gomock.Any()).
					Return(models.ShortenedLink{}, models.ErrUnfound).AnyTimes()

				// Создание записей - ожидаем WithinTx для BatchCreate
				mockStorage.EXPECT().
					ShortenedLinkBatchCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
						for _, url := range urls {
							assert.NotEmpty(t, url.ShortCode)
							assert.False(t, url.CreatedAt.IsZero())
						}
						return urls, nil
					})
			},
			wantResult: []models.ShortenedLink{
				{OriginalURL: "http://url1", UserID: 1},
				{OriginalURL: "http://url2", UserID: 1},
			},
		},
		{
			name: "Часть URL уже существует",
			input: []models.ShortenedLink{
				{OriginalURL: "http://existing", UserID: 1},
				{OriginalURL: "http://new", UserID: 1},
			},
			mockSetup: func() {
				// Возвращаем один существующий URL
				mockStorage.EXPECT().
					ShortenedLinkBatchExists(gomock.Any(), []string{"http://existing", "http://new"}).
					Return([]models.ShortenedLink{
						{OriginalURL: "http://existing", ShortCode: "exist123", UserID: 1},
					}, nil)

				// Проверка существующих shortCode после Генерации shortCode
				mockStorage.EXPECT().
					ShortenedLinkGetByShortKey(gomock.Any(), gomock.Any()).
					Return(models.ShortenedLink{}, models.ErrUnfound).AnyTimes()

				// Создаем только новый URL
				mockStorage.EXPECT().
					ShortenedLinkBatchCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, urls []models.ShortenedLink) ([]models.ShortenedLink, error) {
						assert.Len(t, urls, 1)
						assert.Equal(t, "http://new", urls[0].OriginalURL)
						return urls, nil
					})
			},
			wantResult: []models.ShortenedLink{
				{OriginalURL: "http://existing", ShortCode: "exist123", UserID: 1},
				{OriginalURL: "http://new", UserID: 1},
			},
		},
		{
			name: "Все URL уже существуют",
			input: []models.ShortenedLink{
				{OriginalURL: "http://existing1", UserID: 1},
				{OriginalURL: "http://existing2", UserID: 1},
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkBatchExists(gomock.Any(), []string{"http://existing1", "http://existing2"}).
					Return([]models.ShortenedLink{
						{OriginalURL: "http://existing1", ShortCode: "exist1", UserID: 1},
						{OriginalURL: "http://existing2", ShortCode: "exist2", UserID: 1},
					}, nil)
			},
			wantResult: []models.ShortenedLink{
				{OriginalURL: "http://existing1", ShortCode: "exist1", UserID: 1},
				{OriginalURL: "http://existing2", ShortCode: "exist2", UserID: 1},
			},
			wantErr:     true,
			expectedErr: models.ErrConflict,
		},
		{
			name:        "Пустой пакет",
			input:       []models.ShortenedLink{},
			mockSetup:   func() {},
			wantErr:     true,
			expectedErr: models.ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := service.BatchCreate(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, len(tt.wantResult))

			for i, want := range tt.wantResult {
				assert.Equal(t, want.OriginalURL, got[i].OriginalURL)
				assert.Equal(t, want.UserID, got[i].UserID)
				if want.ShortCode != "" {
					assert.Equal(t, want.ShortCode, got[i].ShortCode)
				} else {
					assert.NotEmpty(t, got[i].ShortCode)
				}
			}
		})
	}
}

func TestURLShortener_GetUserLinks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockURLStorage(ctrl)
	service := NewServiceURLShortener(mockStorage, "http://short")

	tests := []struct {
		name        string
		userID      int64
		mockSetup   func()
		wantLinks   []models.ShortenedLink
		wantErr     bool
		expectedErr string
	}{
		{
			name:   "Успешное получение ссылок пользователя",
			userID: 1,
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkGetBatchByUser(gomock.Any(), int64(1)).
					Return([]models.ShortenedLink{
						{OriginalURL: "http://url1", ShortCode: "code1", UserID: 1},
						{OriginalURL: "http://url2", ShortCode: "code2", UserID: 1},
					}, nil)
			},
			wantLinks: []models.ShortenedLink{
				{OriginalURL: "http://url1", ShortCode: "code1", UserID: 1},
				{OriginalURL: "http://url2", ShortCode: "code2", UserID: 1},
			},
		},
		{
			name:   "Пользователь без ссылок",
			userID: 2,
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkGetBatchByUser(gomock.Any(), int64(2)).
					Return([]models.ShortenedLink{}, nil)
			},
			wantLinks: []models.ShortenedLink{},
		},
		{
			name:        "Невалидный userID",
			userID:      0,
			mockSetup:   func() {},
			wantErr:     true,
			expectedErr: "invalid user ID",
		},
		{
			name:   "Ошибка хранилища",
			userID: 3,
			mockSetup: func() {
				mockStorage.EXPECT().
					ShortenedLinkGetBatchByUser(gomock.Any(), int64(3)).
					Return(nil, models.ErrEmpty)
			},
			wantErr:     true,
			expectedErr: "failed to get user links",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := service.GetUserLinks(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLinks, got)
		})
	}
}

func TestURLShortener_PingDataBase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockURLStorage(ctrl)
	service := NewServiceURLShortener(mockStorage, "http://short")

	tests := []struct {
		name        string
		mockSetup   func()
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Успешный ping",
			mockSetup: func() {
				mockStorage.EXPECT().
					Ping(gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "Ошибка ping",
			mockSetup: func() {
				mockStorage.EXPECT().
					Ping(gomock.Any()).
					Return(fmt.Errorf("connection failed"))
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("database ping failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			err := service.PingDataBase(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
