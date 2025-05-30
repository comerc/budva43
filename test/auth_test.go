package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	config "github.com/comerc/budva43/app/config"
	util "github.com/comerc/budva43/app/util"
	authController "github.com/comerc/budva43/controller/auth"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authService "github.com/comerc/budva43/service/auth"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

func initTelegramDirs(t *testing.T) {
	t.Helper()

	fmt.Println("config.Telegram.UseTestDc:", config.Telegram.UseTestDc)

	currDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("ошибка получения текущей директории: %v", err)
	}
	config.Telegram.LogDirectory = path.Join(currDir, ".data", "telegram", "log")
	config.Telegram.DatabaseDirectory = path.Join(currDir, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = path.Join(currDir, ".data", "telegram", "files")
	var dirs = []string{
		config.Telegram.LogDirectory,
		config.Telegram.DatabaseDirectory,
		config.Telegram.FilesDirectory,
	}
	config.RemoveDirs(dirs...)
	config.MakeDirs(dirs...)
}

func TestAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(func() {
		cancel()
	})

	initTelegramDirs(t)

	var err error

	automator, err := util.NewCLIAutomator()
	require.NoError(t, err)
	t.Cleanup(func() {
		automator.Close()
	})
	go automator.Run()

	telegramRepo := telegramRepo.New()
	err = telegramRepo.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := telegramRepo.Close()
		require.NoError(t, err)
	})

	authService := authService.New(telegramRepo)
	require.NotNil(t, authService)

	err = authService.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := authService.Close()
		require.NoError(t, err)
	})

	time.Sleep(1 * time.Second)

	authController := authController.New(authService)
	require.NotNil(t, authController)

	state := authController.GetState()
	require.NotNil(t, state)
	assert.Equal(t, client.TypeAuthorizationStateWaitPhoneNumber, state.AuthorizationStateType())

	cliTransport := cliTransport.New(
		// reportController,
		authController,
	)
	err = cliTransport.Start(ctx, cancel)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := cliTransport.Close()
		require.NoError(t, err)
	})

	webTransport := webTransport.New(
		// reportController,
		authController,
	)
	err = webTransport.Start(ctx, cancel)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := webTransport.Close()
		require.NoError(t, err)
	})

	var found bool

	found = automator.WaitForOutput(ctx, "Запуск CLI интерфейса", 3*time.Second)
	require.True(t, found, "CLI транспорт не запустился")
	result := automator.WaitForOutput(ctx, "> ", 3*time.Second)
	require.True(t, result, "CLI транспорт не выдал запрос на ввод команды")

	// Проверяем команду help
	err = automator.SendInput("help")
	require.NoError(t, err)
	found = automator.WaitForOutput(ctx, "Доступные команды:", 2*time.Second)
	assert.True(t, found, "Команда help не выдала список команд")

	// Проверяем команду auth
	phoneNumber := config.Telegram.PhoneNumber
	t.Cleanup(func() {
		config.Telegram.PhoneNumber = phoneNumber
	})
	config.Telegram.PhoneNumber = "" // test empty phone number

	err = automator.SendInput("auth")
	require.NoError(t, err)
	found = automator.WaitForOutput(ctx, "Введите номер телефона:", 3*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод номера телефона")

	delimiter := strings.Repeat("*", len(phoneNumber))
	println(delimiter)
	println(util.MaskPhoneNumber(phoneNumber))
	println(delimiter)

	err = automator.SendInput(phoneNumber)
	require.NoError(t, err)
	time.Sleep(3 * time.Second)

	err = automator.SendInput("auth")
	require.NoError(t, err)
	found = automator.WaitForOutput(ctx, "Введите код подтверждения:", 3*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод кода подтверждения")

	err = automator.SendInput("xxx")
	require.NoError(t, err)
	time.Sleep(3 * time.Second)

	target := "http://localhost:7070/api/auth/telegram/state"

	// Отправляем реальный HTTP-запрос к запущенному серверу
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(target)
	require.NoError(t, err, "Ошибка при выполнении HTTP-запроса к %s", target)
	t.Cleanup(func() {
		resp.Body.Close()
	})

	// Проверяем статус ответа
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Статус ответа должен быть 200 OK")

	// Читаем и проверяем тело ответа
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Ошибка при чтении тела ответа")

	responseBody := string(body)
	println(responseBody)
	assert.Contains(t, responseBody, "state_type", "Ответ должен содержать информацию о состоянии авторизации")

	// Проверяем команду exit
	err = automator.SendInput("exit")
	require.NoError(t, err)
	found = automator.WaitForOutput(ctx, "Выход из программы", 3*time.Second)
	assert.True(t, found, "Команда exit не сработала")

	// Проверяем, что контекст был отменен (CLI завершился)
	select {
	case <-ctx.Done():
		// OK, контекст отменен
	case <-time.After(3 * time.Second):
		t.Error("CLI не завершился после команды exit")
	}
}
