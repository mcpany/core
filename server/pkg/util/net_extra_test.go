package util_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListenWithRetry(t *testing.T) {
	// Restore original ListenFunc after tests
	originalListenFunc := util.ListenFunc
	defer func() {
		util.ListenFunc = originalListenFunc
	}()

	t.Run("Success Port 0", func(t *testing.T) {
		util.ListenFunc = originalListenFunc // use real one
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		lis, err := util.ListenWithRetry(ctx, "tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer lis.Close()

		assert.NotNil(t, lis)
		assert.Contains(t, lis.Addr().String(), "127.0.0.1")
	})

	t.Run("Retry on EADDRINUSE with :0", func(t *testing.T) {
		calls := 0
		util.ListenFunc = func(ctx context.Context, network, address string) (net.Listener, error) {
			calls++
			if calls < 3 {
				// Simulate collision
				return nil, fmt.Errorf("bind: address already in use")
			}
			// Succeed
			return originalListenFunc(ctx, network, address)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Must use :0 to trigger retry loop
		lis, err := util.ListenWithRetry(ctx, "tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer lis.Close()

		assert.GreaterOrEqual(t, calls, 3)
	})

	t.Run("Fail immediately on other errors", func(t *testing.T) {
		calls := 0
		util.ListenFunc = func(ctx context.Context, network, address string) (net.Listener, error) {
			calls++
			return nil, fmt.Errorf("permission denied")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := util.ListenWithRetry(ctx, "tcp", "127.0.0.1:0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
		assert.Equal(t, 1, calls, "Should not retry for non-bind errors")
	})

	t.Run("Context Cancelled during retry", func(t *testing.T) {
		util.ListenFunc = func(ctx context.Context, network, address string) (net.Listener, error) {
			return nil, fmt.Errorf("bind: address already in use")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		_, err := util.ListenWithRetry(ctx, "tcp", "127.0.0.1:0")
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		// Should have waited at least a bit
		assert.Greater(t, elapsed, 10*time.Millisecond)
	})
}
