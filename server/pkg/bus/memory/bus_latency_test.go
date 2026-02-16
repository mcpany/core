// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBus_Latency(t *testing.T) {
	// Create a bus with a timeout that is significant but not too long
	bus := New[string]()
	timeout := 200 * time.Millisecond
	bus.publishTimeout = timeout

	// Create a channel to block the subscribers
	block := make(chan struct{})

	// Subscribe 2 handlers that will block
	// This ensures they stop processing messages, causing the channel buffer to fill up
	unsub1 := bus.Subscribe(context.Background(), "latency_test_block", func(_ string) {
		<-block
	})
	defer unsub1()

	unsub2 := bus.Subscribe(context.Background(), "latency_test_block", func(_ string) {
		<-block
	})
	defer unsub2()

	// Channel buffer size is 128 (hardcoded in bus.go)
	// We need to fill the buffer so that the *next* Publish blocks and hits the timeout.
	// 1 message is consumed by the handler (and blocks on 'block' channel)
	// 128 messages fill the buffer
	// So we need 129 messages to saturate the subscriber.
	for i := 0; i < 129; i++ {
		_ = bus.Publish(context.Background(), "latency_test_block", "fill")
	}

	// Now both subscribers are saturated.
	// The next Publish will attempt to send to both.
	// Both should block for 'timeout' duration because their channels are full.

	start := time.Now()
	_ = bus.Publish(context.Background(), "latency_test_block", "slow")
	duration := time.Since(start)

	// Release the blocked handlers
	close(block)

	// Sequential behavior: Wait for Sub 1 (200ms) + Wait for Sub 2 (200ms) = ~400ms
	// Parallel behavior: Wait for both concurrently = ~200ms

	// We expect the optimized version to be close to 200ms.
	// We set a threshold that clearly distinguishes sequential from parallel.
	// 1.5 * timeout = 300ms.
	// If duration > 300ms, it's likely sequential.
	// If duration < 300ms, it's likely parallel.

	t.Logf("Publish duration: %v (timeout: %v)", duration, timeout)

	// Assert that it is fast (Parallel)
	// This assertion is expected to FAIL before the fix is implemented.
	assert.Less(t, duration, 300 * time.Millisecond, "Publish took too long, expected parallel execution")
}
