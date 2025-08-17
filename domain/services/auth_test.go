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

func TestSomething(t *testing.T) {

	// base64.StdEncoding.EncodeToString([]byte()
	assert.True(t, true, "True is true!")
}

func TestAuth_Register(t *testing.T) {
	secretKey := base64.StdEncoding.EncodeToString([]byte("test-secret-key-32-bytes-long!!!"))
	accessExp := 15 * time.Minute

	tests := []struct {
		name      string
		inputUser models.User
		setupMock func(*mocks.MockUserStorage)
		// wantUser    models.User
		// expected    string
		wantErr     bool
		errContains string
	}{
		{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mocks.NewMockUserStorage(ctrl)
			authService, err := NewAuthentication(mock, secretKey, accessExp)
			require.NoError(t, err)
			tt.setupMock(mock)

			gotUser, gotToken, _, err := authService.Register(context.Background(), tt.inputUser)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)
			assert.NotZero(t, gotUser.ID)
			assert.IsType(t, int64(0), gotUser.ID) // типа защита от дурака хз может надо будет убрать
			assert.NotEmpty(t, gotToken)
		})
	}

}
