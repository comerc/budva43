package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	config "github.com/comerc/budva43/config"
	authTelegramController "github.com/comerc/budva43/controller/auth_telegram"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authTelegramService "github.com/comerc/budva43/service/auth_telegram"
)

// TODO: если некорректные параметры клиента, то падает с неопознанной паникой

func TestAuthTelegram(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config.Telegram.UseTestDc = true
	config.Telegram.DatabaseDirectory = "./test/.data/telegram"
	config.Telegram.FilesDirectory = "./test/.data/telegram"
	config.MakeDirs()

	telegramRepo := telegramRepo.New()
	err := telegramRepo.Start(ctx, cancel)
	require.NoError(t, err)
	defer telegramRepo.Stop()

	authTelegramService := authTelegramService.New(telegramRepo)
	require.NotNil(t, authTelegramService)

	time.Sleep(1 * time.Second)

	authTelegramController := authTelegramController.New(authTelegramService)
	require.NotNil(t, authTelegramController)

	state, err := authTelegramController.GetAuthorizationState()
	require.NoError(t, err)
	assert.Equal(t, client.TypeAuthorizationStateWaitPhoneNumber, state.AuthorizationStateType())
}
