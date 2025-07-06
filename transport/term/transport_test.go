package term

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/testing/term_automator"
	realTermRepo "github.com/comerc/budva43/repo/term"
	"github.com/comerc/budva43/transport/term/mocks"
)

func TestTermTransport(t *testing.T) {
	// t.Parallel() // !! нельзя параллелить, т.к. тестирую через term_automator
	// TODO: включить после переделки term_automator

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	var err error

	//TODO: выпилить term_automator из юнит-теста

	var automator *term_automator.TermAutomator
	automator, err = term_automator.NewTermAutomator()
	require.NoError(t, err)
	t.Cleanup(func() {
		automator.Close()
	})

	go automator.Run()

	realTermRepo := realTermRepo.New()

	// termRepo := mocks.NewTermRepo(t)
	// termRepo.EXPECT().ReadLine().
	// 	RunAndReturn(realTermRepo.ReadLine).
	// 	Times(2)

	authService := mocks.NewAuthService(t)
	authService.EXPECT().GetClientDone().
		Return(
			func() <-chan any {
				clientDone := make(chan any)
				close(clientDone)
				return clientDone
			}(),
		)
	authService.EXPECT().GetStatus().
		Return(client.TypeAuthorizationStateWaitPhoneNumber)

	termTransport := New(
		realTermRepo,
		authService,
	)
	termTransport.shutdown = cancel
	go termTransport.runInputLoop(ctx)

	var found bool

	found = automator.WaitForOutput(ctx, ">", 1*time.Second)
	assert.True(t, found, "CLI не выдала приглашение")

	err = automator.SendInput("help")
	require.NoError(t, err)
	found = automator.WaitForOutput(ctx, "Доступные команды:", 1*time.Second)
	assert.True(t, found, "Команда help не выдала список команд")

	err = automator.SendInput("exit")
	require.NoError(t, err)
	select {
	case <-ctx.Done():
		// OK, контекст отменен
	case <-time.After(3 * time.Second):
		t.Error("CLI не завершился после команды exit")
	}
}
