package term

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/transport/term/mocks"
)

func Test(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	synctest.Run(func() {
		status := "TDLib version: 1.0.0 userId: 1234567890"
		clientDone := make(chan any)
		close(clientDone)

		termRepo := mocks.NewTermRepo(t)
		termRepo.EXPECT().Println(status)
		termRepo.EXPECT().Println(">")
		termRepo.EXPECT().ReadLine().Return("exit", nil)
		termRepo.EXPECT().Println("Выход из программы...")

		authService := mocks.NewAuthService(t)
		authService.EXPECT().GetClientDone().Return(clientDone)
		authService.EXPECT().GetStatus().Return(status)

		termTransport := New(
			termRepo,
			authService,
		)
		termTransport.shutdown = cancel

		go termTransport.runInputLoop(ctx)

		select {
		case <-ctx.Done():
			// OK, контекст отменен
		case <-time.After(1 * time.Second):
			t.Error("termTransport не завершился после команды exit")
			cancel()
		}

		t.Cleanup(func() {
			termTransport.Close() // !! вызываем после отмены контекста
		})
	})
}

func TestProcessAuth_WaitPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		passwordHint string
		userInput    string
	}{
		{
			name:         "with hint",
			passwordHint: "test hint",
			userInput:    "password123",
		},
		{
			name:         "without hint",
			passwordHint: "",
			userInput:    "mypassword",
		},
		{
			name:         "with empty hint",
			passwordHint: "",
			userInput:    "secretpass",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			termRepo := mocks.NewTermRepo(t)
			authService := mocks.NewAuthService(t)

			transport := New(termRepo, authService)

			// Создаем состояние ожидания пароля
			passwordState := &client.AuthorizationStateWaitPassword{
				PasswordHint: test.passwordHint,
			}

			// Настраиваем ожидания в зависимости от наличия подсказки
			if test.passwordHint == "" {
				termRepo.EXPECT().Println("Введите пароль: ").Once()
			} else {
				termRepo.EXPECT().Printf("Введите пароль (подсказка: %s): \n", test.passwordHint).Once()
			}

			termRepo.EXPECT().HiddenReadLine().Return(test.userInput, nil).Once()

			// Ожидаем отправку пароля в канал
			inputChan := make(chan string, 1)
			authService.EXPECT().GetInputChan().Return(inputChan).Once()

			// Вызываем метод
			transport.processAuth(passwordState)

			// Проверяем, что пароль был отправлен
			assert.Equal(t, test.userInput, <-inputChan)
		})
	}
}
