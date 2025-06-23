package cli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zelenin/go-tdlib/client"

	"github.com/comerc/budva43/app/testing/cli_automator"
	"github.com/comerc/budva43/transport/cli/mocks"
)

func TestCliTransport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	var err error

	automator, err := cli_automator.NewCLIAutomator()
	require.NoError(t, err)
	t.Cleanup(func() {
		automator.Close()
	})

	go automator.Run()

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

	cliTransport := New(
		authService,
	)
	cliTransport.shutdown = cancel
	go cliTransport.runInputLoop(ctx)

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
