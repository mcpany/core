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
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus/redis"
	"github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

func TestRedisBus_Integration(t *testing.T) {
	t.Skip("This test requires a running redis instance")

	t.Run("should subscribe and receive a message", func(t *testing.T) {
		b := redis.New[map[string]string](&bus.RedisBus{
			Address: "localhost:6379",
		})
		defer b.Close()

		received := make(chan map[string]string, 1)
		handler := func(msg map[string]string) {
			received <- msg
		}

		unsubscribe := b.Subscribe(context.Background(), "test-topic", handler)
		defer unsubscribe()

		time.Sleep(100 * time.Millisecond) // wait for subscription to be active

		err := b.Publish(context.Background(), "test-topic", map[string]string{"key": "value"})
		assert.NoError(t, err)

		select {
		case msg := <-received:
			assert.Equal(t, map[string]string{"key": "value"}, msg)
		case <-time.After(1 * time.Second):
			t.Fatal("timed out waiting for message")
		}
	})
}
