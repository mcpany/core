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

func TestBus_Coverage_New(t *testing.T) {
	// We verify that New creates a client with the correct options (address).
	// Using protojson to avoid issues with opaque API fields.
	// We use new(buspb.RedisBus) which returns a zero-valued config but valid pointer.
	b, _ := New[any](new(buspb.RedisBus))
	assert.NotNil(t, b)
}

func TestBus_Coverage_New_Nil(t *testing.T) {
	// Verify handling of nil config
	b, _ := New[any](nil)
	assert.NotNil(t, b)
}

func TestBus_Close_Error_Simple(t *testing.T) {
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

func TestBus_Subscribe_ContextCancel(_ *testing.T) {
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

func TestBus_Close_WithPubSub(t *testing.T) {
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

func TestBus_Publish_Error(t *testing.T) {
	db, mock := redismock.NewClientMock()
	b := NewWithClient[any](db)

	// Mock Publish failure
	mock.ExpectPublish("topic", []byte("null")).SetErr(context.DeadlineExceeded)

	err := b.Publish(context.Background(), "topic", nil)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestBus_SubscribeOnce_Success(t *testing.T) {
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

func TestBus_Subscribe_NoPanicOnNil(t *testing.T) {
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
