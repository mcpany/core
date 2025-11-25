/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func TestRedisBus_Close(t *testing.T) {
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

func TestRedisBus_Publish(t *testing.T) {
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
			defer b.Close()

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
		defer b.Close()

		err := b.Publish(context.Background(), "test-topic", make(chan int))
		assert.Error(t, err)
	})
}

func TestRedisBus_Subscribe(t *testing.T) {
	t.Parallel()
	db, _ := redismock.NewClientMock()
	b := redis.NewWithClient[string](db)
	defer b.Close()
	t.Run("should panic if handler is nil", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			b.Subscribe(context.Background(), "test-topic", nil)
		})
	})
}

func TestRedisBus_SubscribeOnce(t *testing.T) {
	t.Parallel()
	db, _ := redismock.NewClientMock()
	b := redis.NewWithClient[string](db)
	defer b.Close()
	t.Run("should panic if handler is nil", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			b.SubscribeOnce(context.Background(), "test-topic", nil)
		})
	})
}
