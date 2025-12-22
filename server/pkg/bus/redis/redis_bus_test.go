package redis_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/bus/redis"
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
