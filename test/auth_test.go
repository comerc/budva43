package test

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/test/cli_automator"
	"github.com/comerc/budva43/app/util"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authService "github.com/comerc/budva43/service/auth"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

func TestAuth(t *testing.T) {
	// t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	require.True(t, config.Telegram.UseTestDc)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// X := rand.Intn(3) + 1
	X := 2 // этот работает стабильнее
	Y := rand.Perm(10)[:4]
	phoneNumber := fmt.Sprintf("99966%d%d%d%d%d", X, Y[0], Y[1], Y[2], Y[3])
	// delimiter := strings.Repeat("*", len(phoneNumber))
	// println(delimiter)
	// println(util.MaskPhoneNumber(phoneNumber))
	// println(delimiter)

	var err error

	automator, err := cli_automator.NewCLIAutomator()
	require.NoError(t, err)
	t.Cleanup(func() {
		automator.Close()
	})

	go automator.Run()

	currDir := util.GetCurrDir()

	options := telegramRepo.Options{
		DatabaseDirectory: filepath.Join(currDir, ".data", "telegram", "db"),
		FilesDirectory:    filepath.Join(currDir, ".data", "telegram", "files"),
	}

	var dirs = []string{
		options.DatabaseDirectory,
		options.FilesDirectory,
	}
	for _, dir := range dirs {
		util.RemoveDir(dir)
		util.MakeDir(dir)
	}

	telegramRepo := telegramRepo.New().WithOptions(options)
	err = telegramRepo.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := telegramRepo.Close()
		require.NoError(t, err)
	})

	authService := authService.New(telegramRepo)

	err = authService.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := authService.Close()
		require.NoError(t, err)
	})

	cliTransport := cliTransport.New(
		authService,
	).WithPhoneNumber("")
	err = cliTransport.Start(ctx, cancel)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := cliTransport.Close()
		require.NoError(t, err)
	})

	webTransport := webTransport.New(
		authService,
	)
	err = webTransport.Start(ctx, cancel)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := webTransport.Close()
		require.NoError(t, err)
	})

	var found bool

	found = automator.WaitForOutput(ctx, "Введите номер телефона:", 3*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод номера телефона")

	err = automator.SendInput(phoneNumber)
	require.NoError(t, err)

	found = automator.WaitForOutput(ctx, "Введите код подтверждения:", 5*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод кода подтверждения")

	code := strings.Repeat(fmt.Sprintf("%d", X), 5)
	err = automator.SendInput(code)
	require.NoError(t, err)

	target := fmt.Sprintf("http://%s/api/auth/telegram/state",
		net.JoinHostPort(config.Web.Host, config.Web.Port))

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
	assert.Equal(t, `{"state_type":"authorizationStateWaitCode"}`+"\n", responseBody)

	// после отправки кода подтверждения не поменялся статус авторизации
	// авторизация для UseTestDc не работает: https://github.com/tdlib/td/issues/3361

	cancel()
}
