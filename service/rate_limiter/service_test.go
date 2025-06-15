package rate_limiter

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterService(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	synctest.Run(func() {
		rateLimiter := New()

		dstChatId := int64(123)

		var elapsed time.Duration
		start := time.Now()

		rateLimiter.WaitForForward(ctx, dstChatId)
		elapsed = time.Since(start)
		assert.Equal(t, 0*time.Second, elapsed, "Первый вызов не должен ждать")

		rateLimiter.WaitForForward(ctx, dstChatId)
		elapsed = time.Since(start)
		assert.Equal(t, 3*time.Second, elapsed, "Второй вызов должен ждать 3 секунды")

		cancel()
	})
}
