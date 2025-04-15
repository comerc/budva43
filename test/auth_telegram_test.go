package test

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	config "github.com/comerc/budva43/config"
	authTelegramController "github.com/comerc/budva43/controller/auth_telegram"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authTelegramService "github.com/comerc/budva43/service/auth_telegram"
	cliTransport "github.com/comerc/budva43/transport/cli"
	util "github.com/comerc/budva43/util"
)

func initTelegram(t *testing.T) {
	t.Helper()

	currDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("ошибка получения текущей директории: %v", err)
	}
	config.Telegram.UseTestDc = true
	config.Telegram.DatabaseDirectory = path.Join(currDir, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = path.Join(currDir, ".data", "telegram", "files")
	var dirs = []string{
		config.Telegram.DatabaseDirectory,
		config.Telegram.FilesDirectory,
	}
	config.RemoveDirs(dirs...)
	config.MakeDirs(dirs...)
}

func TestAuthTelegram(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	initTelegram(t)

	var err error

	automator, err := util.NewCLIAutomator()
	require.NoError(t, err)
	defer automator.Stop()
	go automator.Run()

	telegramRepo := telegramRepo.New()
	err = telegramRepo.Start(ctx, cancel)
	require.NoError(t, err)
	defer telegramRepo.Stop()

	authTelegramService := authTelegramService.New(telegramRepo)
	require.NotNil(t, authTelegramService)

	time.Sleep(1 * time.Second) // TODO: dirty hack

	authTelegramController := authTelegramController.New(authTelegramService)
	require.NotNil(t, authTelegramController)

	state, err := authTelegramController.GetAuthorizationState()
	require.NoError(t, err)
	assert.Equal(t, client.TypeAuthorizationStateWaitPhoneNumber, state.AuthorizationStateType())

	cliTransport := cliTransport.New(
		nil, // messageController,
		nil, // forwardController,
		nil, // reportController,
		authTelegramController,
	)
	err = cliTransport.Start(ctx, cancel)
	require.NoError(t, err)
	defer cliTransport.Stop()

	var found bool

	found = automator.WaitForOutput("Запуск CLI интерфейса", 3*time.Second)
	require.True(t, found, "CLI транспорт не запустился")
	result := automator.WaitForOutput("> ", 3*time.Second)
	require.True(t, result, "CLI транспорт не выдал запрос на ввод команды")

	// Проверяем команду help
	automator.SendInput("help")
	found = automator.WaitForOutput("Доступные команды:", 2*time.Second)
	assert.True(t, found, "Команда help не выдала список команд")

	// Проверяем команду auth
	phoneNumber := config.Telegram.PhoneNumber
	defer func() {
		config.Telegram.PhoneNumber = phoneNumber
	}()
	config.Telegram.PhoneNumber = "" // test empty phone number

	err = automator.SendInput("auth")
	require.NoError(t, err)
	found = automator.WaitForOutput("Введите номер телефона:", 3*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод номера телефона")

	delimiter := strings.Repeat("*", len(phoneNumber))
	println(delimiter)
	println(util.MaskPhoneNumber(phoneNumber))
	println(delimiter)

	time.Sleep(3 * time.Second)
	err = automator.SendInput(phoneNumber)
	require.NoError(t, err)

	automator.SendInput("auth")
	found = automator.WaitForOutput("Введите код подтверждения:", 3*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод кода подтверждения")

	time.Sleep(3 * time.Second)
	err = automator.SendInput("xxx")
	require.NoError(t, err)

	// Проверяем команду exit
	err = automator.SendInput("exit")
	require.NoError(t, err)
	found = automator.WaitForOutput("Выход из программы", 3*time.Second)
	assert.True(t, found, "Команда exit не сработала")

	// Проверяем, что контекст был отменен (CLI завершился)
	select {
	case <-ctx.Done():
		// OK, контекст отменен
	case <-time.After(3 * time.Second):
		t.Error("CLI не завершился после команды exit")
	}
}
