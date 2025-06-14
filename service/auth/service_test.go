package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/service/auth/mocks"
)

func TestAuthService_GetStatus(t *testing.T) {
	t.Parallel()

	telegramRepo := mocks.NewTelegramRepo(t)
	telegramRepo.EXPECT().GetOption(&client.GetOptionRequest{Name: "version"}).
		Return(&client.OptionValueString{Value: "1.2.3"}, nil)
	telegramRepo.EXPECT().GetMe().
		Return(&client.User{Id: 123}, nil)
	service := New(telegramRepo)

	status := service.GetStatus()
	assert.Equal(t, "TDLib version: 1.2.3 userId: 123", status)
}
