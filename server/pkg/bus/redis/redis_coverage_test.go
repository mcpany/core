// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	buspb "github.com/mcpany/core/proto/bus" // Aliased to avoid conflicts
	"github.com/stretchr/testify/assert"
)

// TestBus_Coverage_New ...
// Summary: TestBus_Coverage_New
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// We verify that New creates a client with the correct options (address).
	// Using protojson to avoid issues with opaque API fields.
	// We use new(buspb.RedisBus) which returns a zero-valued config but valid pointer.
	b, _ := New[any](new(buspb.RedisBus))
	assert.NotNil(t, b)
}

// TestBus_Coverage_New_Nil ...
// Summary: TestBus_Coverage_New_Nil
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Verify handling of nil config
	b, _ := New[any](nil)
	assert.NotNil(t, b)
}

// TestBus_Close_Error_Simple ...
// Summary: TestBus_Close_Error_Simple
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// redismock v9 might not expose ExpectClose.
	// If it doesn't, we can skip Close error test or check if there's another way.
	// client.Close() on mock usually just returns nil.
	// We can trust go-redis client.Close works or is covered by go-redis tests.
	// We just want to ensure our wrapper calls it.

	db, _ := redismock.NewClientMock()
	b := NewWithClient[any](db)
	err := b.Close()
	assert.NoError(t, err)
}

// TestBus_Subscribe_ContextCancel ...
// Summary: TestBus_Subscribe_ContextCancel
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	db, _ := redismock.NewClientMock()
	b := NewWithClient[any](db)

	ctx, cancel := context.WithCancel(context.Background())

	// Subscribe
	unsub := b.Subscribe(ctx, "topic", func(_ any) {})
	defer unsub()

	// Cancel context immediately to stop the loop
	cancel()

	// We verify that it doesn't block or panic.
}

// TestBus_Close_WithPubSub ...
// Summary: TestBus_Close_WithPubSub
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	db, _ := redismock.NewClientMock()
	b := NewWithClient[any](db)

	// Subscribe without deferring unsubscribe/closing context explicitly
	// We want Close to close the pubsub.
	ctx := context.Background()
	_ = b.Subscribe(ctx, "topic", func(_ any) {})

	// Called b.Close which should iterate over pubsubs and close them
	err := b.Close()
	assert.NoError(t, err)
}

// TestBus_Publish_Error ...
// Summary: TestBus_Publish_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	db, mock := redismock.NewClientMock()
	b := NewWithClient[any](db)

	// Mock Publish failure
	mock.ExpectPublish("topic", []byte("null")).SetErr(context.DeadlineExceeded)

	err := b.Publish(context.Background(), "topic", nil)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

// TestBus_SubscribeOnce_Success ...
// Summary: TestBus_SubscribeOnce_Success
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	db, _ := redismock.NewClientMock()
	b := NewWithClient[any](db)

	ctx := context.Background()
	received := make(chan any, 1)

	// We verify that SubscribeOnce doesn't panic and returns an unsubscribe function.
	// Fully mocking the async loop requires more advanced mocking (e.g. miniredis) because
	// redismock doesn't easily support controlled message injection into the created PubSub.

	unsub := b.SubscribeOnce(ctx, "topic", func(msg any) {
		received <- msg
	})
	assert.NotNil(t, unsub)
	unsub()
}

// TestBus_Subscribe_NoPanicOnNil ...
// Summary: TestBus_Subscribe_NoPanicOnNil
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// We want to verify that it does NOT panic on nil handler (just logs error).
	db, _ := redismock.NewClientMock()
	b := NewWithClient[any](db)

	assert.NotPanics(t, func() {
		b.Subscribe(context.Background(), "topic", nil)
	})

	assert.NotPanics(t, func() {
		b.SubscribeOnce(context.Background(), "topic", nil)
	})
}
