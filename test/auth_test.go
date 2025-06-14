package test

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/comerc/budva43/app/config"
	"github.com/comerc/budva43/app/test_util"
	"github.com/comerc/budva43/app/util"
	telegramRepo "github.com/comerc/budva43/repo/telegram"
	authService "github.com/comerc/budva43/service/auth"
	cliTransport "github.com/comerc/budva43/transport/cli"
	webTransport "github.com/comerc/budva43/transport/web"
)

func TestMain(m *testing.M) {
	currDir := util.GetCurrDir()
	config.Telegram.DatabaseDirectory = filepath.Join(currDir, ".data", "telegram", "db")
	config.Telegram.FilesDirectory = filepath.Join(currDir, ".data", "telegram", "files")

	var dirs = []string{
		config.Telegram.DatabaseDirectory,
		config.Telegram.FilesDirectory,
	}
	for _, dir := range dirs {
		util.RemoveDir(dir)
		util.MakeDir(dir)
	}
	os.Exit(m.Run())
}

func TestAuth(t *testing.T) {
	// t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	require.True(t, config.Telegram.UseTestDc)

	// Проверяем команду auth
	oldPhoneNumber := config.Telegram.PhoneNumber
	t.Cleanup(func() {
		config.Telegram.PhoneNumber = oldPhoneNumber
	})
	config.Telegram.PhoneNumber = "" // test empty phone number

	var err error

	automator, err := test_util.NewCLIAutomator()
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

	cliTransport := cliTransport.New(
		authService,
	)
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

	// X := rand.Intn(3) + 1
	X := 2 // этот работает стабильнее
	Y := rand.Perm(10)[:4]
	newPhoneNumber :=
		fmt.Sprintf("99966%d%d%d%d%d", X, Y[0], Y[1], Y[2], Y[3])
	delimiter := strings.Repeat("*", len(newPhoneNumber))
	println(delimiter)
	println(util.MaskPhoneNumber(newPhoneNumber))
	println(delimiter)
	err = automator.SendInput(newPhoneNumber)
	require.NoError(t, err)

	found = automator.WaitForOutput(ctx, "Введите код подтверждения:", 11*time.Second)
	assert.True(t, found, "Команда auth не выдала запрос на ввод кода подтверждения")

	code := strings.Repeat(fmt.Sprintf("%d", X), 5)
	err = automator.SendInput(code)
	require.NoError(t, err)

	// TODO: логин для UseTestDc пока не работает https://github.com/tdlib/td/issues/3361

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

	cancel()

	// // Проверяем команду help
	// err = automator.SendInput("help")
	// require.NoError(t, err)
	// found = automator.WaitForOutput(ctx, "Доступные команды:", 2*time.Second)
	// assert.True(t, found, "Команда help не выдала список команд")

	// // Проверяем команду exit
	// err = automator.SendInput("exit")
	// require.NoError(t, err)
	// select {
	// case <-ctx.Done():
	// 	// OK, контекст отменен
	// case <-time.After(3 * time.Second):
	// 	t.Error("CLI не завершился после команды exit")
	// }
}
