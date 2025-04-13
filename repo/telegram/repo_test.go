package telegram

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/comerc/budva43/config"
)

func TestRepo_Start(t *testing.T) {

	config.Telegram.UseTestDc = true

	// Выведем информацию о путях для отладки
	fmt.Println("Telegram Database Directory:", config.Telegram.DatabaseDirectory)
	fmt.Println("Telegram Files Directory:", config.Telegram.FilesDirectory)

	fmt.Println("UseTestDc:", config.Telegram.UseTestDc)

	fmt.Println("repo started and stopped")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := New()
	err := repo.Start(ctx, cancel)
	if err != nil {
		t.Fatalf("failed to start repo: %v", err)
	}

	time.Sleep(5 * time.Second)

	c := repo.GetClient()

	assert.Nil(t, c)

	repo.Stop()
}
