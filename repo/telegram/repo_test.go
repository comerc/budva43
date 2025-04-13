package telegram

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestRepo_Start(t *testing.T) {
	slog.Info("repo started and stopped")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := New()
	err := repo.Start(ctx, cancel)
	if err != nil {
		t.Fatalf("failed to start repo: %v", err)
	}

	time.Sleep(5 * time.Second)

	c := repo.GetClient()
	slog.Info("client", "client", c)

	repo.Stop()

}
