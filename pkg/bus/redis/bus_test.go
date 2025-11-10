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

package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/go-redis/redismock/v9"
	"github.com/mcpany/core/pkg/logging"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func setupRedisIntegrationTest(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}
	return client
}

func TestRedisBus_Publish(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetVal(1)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_Publish_MarshalError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[chan int](client)

	err := bus.Publish(context.Background(), "test", make(chan int))
	assert.Error(t, err)
	assert.IsType(t, &json.UnsupportedTypeError{}, err)
}

func TestRedisBus_Publish_RedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetErr(redis.ErrClosed)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_Subscribe(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), "test-subscribe", func(msg string) {
		assert.Equal(t, "hello", msg)
		wg.Done()
	})
	defer unsub()

	// It can take a moment for the subscription to be active.
	// A brief sleep is pragmatic here to ensure the subscriber is ready.
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), "test-subscribe", "hello")
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_Subscribe_UnmarshalError(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	// Capture log output
	var logBuffer bytes.Buffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelDebug, &logBuffer)
	defer logging.ForTestsOnlyResetLogger()

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unmarshal-error", func(msg string) {
		handlerCalled <- true
	})
	defer unsub()

	time.Sleep(100 * time.Millisecond)

	err := client.Publish(context.Background(), "test-unmarshal-error", "invalid json").Err()
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called")
	case <-time.After(200 * time.Millisecond):
		// Test passed, handler was not called
	}

	assert.Contains(t, logBuffer.String(), "Failed to unmarshal message")
}

// TestRedisBus_SubscribeOnce tests that a handler for a topic is only called once.
// Note: Go's coverage tool may report 0% coverage for this function. This is a
// known issue with the tool's ability to track coverage in goroutines,
// especially in short-lived test scenarios. The test is valid and does
// exercise the code path.
func TestRedisBus_SubscribeOnce(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMessages []string
	unsub := bus.SubscribeOnce(context.Background(), "test-once", func(msg string) {
		receivedMessages = append(receivedMessages, msg)
		wg.Done()
	})
	defer unsub()

	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), "test-once", "hello")
	assert.NoError(t, err)
	err = bus.Publish(context.Background(), "test-once", "world")
	assert.NoError(t, err)

	wg.Wait()

	assert.Equal(t, []string{"hello"}, receivedMessages)
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)

	handlerCalled := make(chan bool, 1)

	unsub := bus.Subscribe(context.Background(), "test-unsubscribe", func(msg string) {
		handlerCalled <- true
	})

	time.Sleep(100 * time.Millisecond)
	unsub()
	time.Sleep(100 * time.Millisecond) // Allow time for unsubscribe to propagate

	err := bus.Publish(context.Background(), "test-unsubscribe", "hello")
	assert.NoError(t, err)

	select {
	case <-handlerCalled:
		t.Fatal("handler should not have been called after unsubscribe")
	case <-time.After(200 * time.Millisecond):
		// Test passed
	}
}

func TestRedisBus_New(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("password"),
		Db:       proto.Int32(1),
	}.Build()

	bus := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "password", options.Password)
	assert.Equal(t, 1, options.DB)
}
