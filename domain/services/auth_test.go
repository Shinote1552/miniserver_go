package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"
	"urlshortener/domain/models"
	"urlshortener/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
DECRYPTION

Register(ctx context.Context, user models.User) (models.User, string, time.Time, error)
ValidateAndGetUser(ctx context.Context, jwtToken string) (models.User, error)

Эти два метода тесно связаны, чтобы работало одно нужно другое тоже.
Надо ли в таких случаях Обьединить тесты этих двух методов в один?
Пока что написал с разделением.
*/

func TestAuth_Register(t *testing.T) {
	secretKey := base64.StdEncoding.EncodeToString([]byte("test-secret-key-32-bytes-long!!!"))
	accessExp := 15 * time.Minute

	tests := []struct {
		name        string
		inputUser   models.User
		mockSetup   func(*mocks.MockUserStorage, models.User)
		wantUserID  int64
		wantToken   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "Успешная регистрация нового пользователя",
			inputUser: models.User{},

			mockSetup: func(m *mocks.MockUserStorage, expected models.User) {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, u models.User) (models.User, error) {
						assert.False(t, u.CreatedAt.IsZero(), "CreatedAt не должно быть нулевым")
						return models.User{
							ID:        1,
							CreatedAt: u.CreatedAt,
						}, nil
					})
				m.EXPECT().
					UserGetByID(gomock.Any(), int64(1)).
					Return(models.User{ID: 1}, nil)
			},
			wantUserID: 1,
			wantToken:  true,
		},
		{
			name:      "Конфликт при создании пользователя",
			inputUser: models.User{},

			mockSetup: func(m *mocks.MockUserStorage, expected models.User) {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, u models.User) (models.User, error) {
						assert.False(t, u.CreatedAt.IsZero(), "CreatedAt не должно быть нулевым")
						return models.User{}, models.ErrConflict
					})
			},
			wantErr:     true,
			expectedErr: models.ErrConflict,
		},
		{
			name:      "Ошибка пустого хранилища",
			inputUser: models.User{},

			mockSetup: func(m *mocks.MockUserStorage, expected models.User) {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, u models.User) (models.User, error) {
						assert.False(t, u.CreatedAt.IsZero(), "CreatedAt не должно быть нулевым")
						return models.User{}, models.ErrEmpty
					})
			},
			wantErr:     true,
			expectedErr: models.ErrEmpty,
		},
		{
			name: "Пользователь с предустановленным CreatedAt",
			inputUser: models.User{
				CreatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},

			mockSetup: func(m *mocks.MockUserStorage, expected models.User) {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, u models.User) (models.User, error) {
						assert.False(t, u.CreatedAt.IsZero(), "CreatedAt не должно быть нулевым")
						assert.True(t, u.CreatedAt.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)), "CreatedAt слишком старое")
						assert.True(t, u.CreatedAt.Before(time.Now().Add(time.Minute)), "CreatedAt в будущем")
						return models.User{
							ID:        2,
							CreatedAt: u.CreatedAt,
						}, nil
					})
				m.EXPECT().
					UserGetByID(gomock.Any(), int64(2)).
					Return(models.User{ID: 2}, nil)
			},
			wantUserID: 2,
			wantToken:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockUserStorage(ctrl)
			auth, err := NewAuthentication(mockStorage, secretKey, accessExp)
			require.NoError(t, err)

			tt.mockSetup(mockStorage, tt.inputUser)

			gotUser, gotToken, _, err := auth.Register(context.Background(), tt.inputUser)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, gotUser.ID)
			assert.False(t, gotUser.CreatedAt.IsZero())

			if tt.wantToken {
				assert.NotEmpty(t, gotToken)

				validatedUser, err := auth.ValidateAndGetUser(context.Background(), gotToken)
				require.NoError(t, err, "Токен должен быть валидным")
				assert.Equal(t, tt.wantUserID, validatedUser.ID, "Токен должен содержать правильный userID")
			}
		})
	}
}

func TestAuth_ValidateAndGetUser(t *testing.T) {
	secretKey := base64.StdEncoding.EncodeToString([]byte("test-secret-key-32-bytes-long!!!"))
	accessExp := 15 * time.Minute

	tests := []struct {
		name        string
		setup       func(*mocks.MockUserStorage) string
		wantUserID  int64
		wantErr     bool
		expectedErr error
	}{
		{
			name: "Успешная валидация токена",

			setup: func(m *mocks.MockUserStorage) string {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					Return(models.User{ID: 1}, nil)
				m.EXPECT().
					UserGetByID(gomock.Any(), int64(1)).
					Return(models.User{ID: 1}, nil)

				auth, _ := NewAuthentication(m, secretKey, accessExp)
				_, token, _, err := auth.Register(context.Background(), models.User{})
				require.NoError(t, err)
				return token
			},
			wantUserID: 1,
		},
		{
			name: "Пользователь не найден",

			setup: func(m *mocks.MockUserStorage) string {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					Return(models.User{ID: 1}, nil)
				m.EXPECT().
					UserGetByID(gomock.Any(), int64(1)).
					Return(models.User{}, models.ErrUnfound)

				auth, _ := NewAuthentication(m, secretKey, accessExp)
				_, token, _, err := auth.Register(context.Background(), models.User{})
				require.NoError(t, err)
				return token
			},
			wantErr:     true,
			expectedErr: models.ErrUnfound,
		},
		{
			name: "Невалидный токен",

			setup: func(m *mocks.MockUserStorage) string {
				return "invalid.token.here"
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("failed to parse token"),
		},
		{
			name: "Просроченный токен",

			setup: func(m *mocks.MockUserStorage) string {
				m.EXPECT().
					UserCreate(gomock.Any(), gomock.Any()).
					Return(models.User{ID: 1}, nil)

				// Создаем auth сервис с отрицательным временем жизни токена
				expiredAuth, err := NewAuthentication(m, secretKey, -1*time.Hour)
				require.NoError(t, err)

				_, token, _, err := expiredAuth.Register(context.Background(), models.User{})
				require.NoError(t, err)
				return token
			},
			wantErr: true,
			// nil чтобы конкретный error не проверял, а лишь факт того что есть ошибка а какая неважно
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockUserStorage(ctrl)
			auth, err := NewAuthentication(mockStorage, secretKey, accessExp)
			require.NoError(t, err)

			token := tt.setup(mockStorage)
			gotUser, err := auth.ValidateAndGetUser(context.Background(), token)

			if tt.wantErr {
				require.Error(t, err)

				if tt.expectedErr != nil {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUserID, gotUser.ID)
		})
	}
}
