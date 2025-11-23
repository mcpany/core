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
	"context"
	"sync"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRedisBus_New_NilConfigExtra(t *testing.T) {
	var bus *RedisBus[string]
	assert.NotPanics(t, func() {
		bus = New[string](nil)
	})
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "", options.Password)
	assert.Equal(t, 0, options.DB)
}

func TestRedisBus_New_PartialConfigExtra(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6381"),
	}.Build()

	bus := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6381", options.Addr)
	assert.Equal(t, "", options.Password)
	assert.Equal(t, 0, options.DB)
}

func TestRedisBus_Subscribe_NilMessage(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[*string](client)
	topic := "test-nil-message"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg *string) {
		assert.Nil(t, msg)
		wg.Done()
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs := client.PubSubNumSub(context.Background(), topic).Val()
		return len(subs) > 0 && subs[topic] == 1
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Publish a "null" JSON payload, which will be unmarshaled to a nil pointer.
	err := client.Publish(context.Background(), topic, "null").Err()
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_New_ConnectionFailure(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:9999"), // Use a non-existent port
	}.Build()

	bus := New[string](redisBus)
	err := bus.client.Ping(context.Background()).Err()
	assert.Error(t, err, "Expected an error when connecting to a non-existent Redis server")
}

func TestRedisBus_Subscribe_ReceiveError(t *testing.T) {
	client := setupRedisIntegrationTest(t)
	bus := NewWithClient[string](client)
	topic := "test-receive-error"

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), topic, func(msg string) {
		// This handler should not be called.
		t.Error("handler called unexpectedly")
	})
	defer unsub()

	require.Eventually(t, func() bool {
		subs, err := client.PubSubNumSub(context.Background(), topic).Result()
		require.NoError(t, err)
		if val, ok := subs[topic]; ok {
			return val == 1
		}
		return false
	}, 1*time.Second, 10*time.Millisecond, "subscriber did not appear")

	// Close the underlying connection to simulate a receive error.
	bus.client.Close()

	// Allow some time for the error to be processed.
	time.Sleep(100 * time.Millisecond)

	// We expect the handler not to be called, so we don't wait for a WaitGroup.
	// Instead, we just wait a bit to see if the handler is called.
	time.Sleep(100 * time.Millisecond)
}
