package test

import (
	"context"
	"log"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedis(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"redis:7",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		redis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
	)
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	})
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
}
