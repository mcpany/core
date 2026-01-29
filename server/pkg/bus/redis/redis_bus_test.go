package redis_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/server/pkg/bus/redis"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithClient(t *testing.T) {
	t.Parallel()
	db, _ := redismock.NewClientMock()
	b := redis.NewWithClient[any](db)
	require.NotNil(t, b)
}

func TestBus_Close(t *testing.T) {
	t.Parallel()

	t.Run("should close the client and pubsubs successfully", func(t *testing.T) {
		t.Parallel()

		db, _ := redismock.NewClientMock()
		b := redis.NewWithClient[any](db)

		// Calling Close should not panic
		err := b.Close()
		assert.NoError(t, err)
	})
}

func TestBus_Publish(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		setupMock   func(mock redismock.ClientMock)
		expectError bool
	}

	testCases := []testCase{
		{
			name: "should publish message successfully",
			setupMock: func(mock redismock.ClientMock) {
				payload, _ := json.Marshal(map[string]string{"key": "value"})
				mock.ExpectPublish("test-topic", payload).SetVal(1)
			},
			expectError: false,
		},
		{
			name: "should return an error when publish fails",
			setupMock: func(mock redismock.ClientMock) {
				payload, _ := json.Marshal(map[string]string{"key": "value"})
				mock.ExpectPublish("test-topic", payload).SetErr(errors.New("publish error"))
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock := redismock.NewClientMock()
			b := redis.NewWithClient[map[string]string](db)
			t.Cleanup(func() { _ = b.Close() })

			tc.setupMock(mock)

			err := b.Publish(context.Background(), "test-topic", map[string]string{"key": "value"})

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}

	t.Run("should return an error for non-marshallable message", func(t *testing.T) {
		t.Parallel()

		db, _ := redismock.NewClientMock()
		// Use a channel, which cannot be marshalled to JSON
		b := redis.NewWithClient[chan int](db)
		t.Cleanup(func() { _ = b.Close() })

		err := b.Publish(context.Background(), "test-topic", make(chan int))
		assert.Error(t, err)
	})
}

func TestBus_Subscribe(t *testing.T) {
	t.Parallel()
	db, _ := redismock.NewClientMock()
	b := redis.NewWithClient[string](db)
	t.Cleanup(func() { _ = b.Close() })
	t.Run("should not panic if handler is nil", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			b.Subscribe(context.Background(), "test-topic", nil)
		})
	})
}

func TestBus_SubscribeOnce(t *testing.T) {
	t.Parallel()
	db, _ := redismock.NewClientMock()
	b := redis.NewWithClient[string](db)
	t.Cleanup(func() { _ = b.Close() })
	t.Run("should not panic if handler is nil", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			b.SubscribeOnce(context.Background(), "test-topic", nil)
		})
	})
}

func TestBus_SubscribeOnce_Success(t *testing.T) {
	t.Parallel()
	client, _ := redismock.NewClientMock()
	bus := redis.NewWithClient[string](client)

	// Just calling it to cover setup lines.
	// We don't verify subscription because redismock might not support it fully without expectations.
	// But calling it covers the non-panic path.
	handler := func(_ string) {}
	_ = bus.SubscribeOnce(context.Background(), "test-topic", handler)
}

func TestBus_Subscribe_Miniredis(t *testing.T) {
	// Start miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redisclient.NewClient(&redisclient.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	bus := redis.NewWithClient[string](client)
	defer bus.Close() // In reality this closes client too, but miniredis handles connection closing gracefully.

	t.Run("Success", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		var receivedMsg string

		unsubscribe := bus.Subscribe(context.Background(), "test-topic-1", func(msg string) {
			receivedMsg = msg
			wg.Done()
		})
		defer unsubscribe()

		// Wait for subscription to be registered
		time.Sleep(50 * time.Millisecond)

		err := bus.Publish(context.Background(), "test-topic-1", "hello")
		require.NoError(t, err)

		wg.Wait()
		assert.Equal(t, "hello", receivedMsg)
	})

	t.Run("UnmarshalError", func(t *testing.T) {
		// We want to verify that an invalid JSON payload is ignored/logged and doesn't crash the subscriber.
		// Since we can't easily assert on logs, we verify that subsequent valid messages are still received.

		var wg sync.WaitGroup
		wg.Add(1)
		var receivedMsg string

		unsubscribe := bus.Subscribe(context.Background(), "test-topic-2", func(msg string) {
			receivedMsg = msg
			wg.Done()
		})
		defer unsubscribe()

		time.Sleep(50 * time.Millisecond)

		// Publish invalid JSON directly via client to bypass Bus.Publish marshaling
		err := client.Publish(context.Background(), "test-topic-2", "{invalid-json").Err()
		require.NoError(t, err)

		// Publish valid message
		err = bus.Publish(context.Background(), "test-topic-2", "valid")
		require.NoError(t, err)

		wg.Wait()
		assert.Equal(t, "valid", receivedMsg)
	})

	t.Run("HandlerPanic", func(t *testing.T) {
		// Verify that a panic in the handler is recovered and subscription continues.
		var wg sync.WaitGroup
		wg.Add(2) // 1 for panic message, 1 for subsequent success

		// We use a counter to differentiate the calls
		callCount := 0
		var receivedMsg string

		unsubscribe := bus.Subscribe(context.Background(), "test-topic-3", func(msg string) {
			callCount++
			if msg == "panic" {
				wg.Done()
				panic("oops")
			}
			receivedMsg = msg
			wg.Done()
		})
		defer unsubscribe()

		time.Sleep(50 * time.Millisecond)

		err := bus.Publish(context.Background(), "test-topic-3", "panic")
		require.NoError(t, err)

		// Wait a bit for panic to happen and recover
		time.Sleep(50 * time.Millisecond)

		err = bus.Publish(context.Background(), "test-topic-3", "recover")
		require.NoError(t, err)

		wg.Wait()
		assert.Equal(t, "recover", receivedMsg)
		assert.Equal(t, 2, callCount)
	})
}
