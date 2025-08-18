package services

import (
	"context"
	"encoding/base64"
	"testing"
	"time"
	"urlshortener/domain/models"
	"urlshortener/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			}
		})
	}
}
