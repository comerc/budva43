package term

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

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
