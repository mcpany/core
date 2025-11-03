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

package nats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/stretchr/testify/assert"
)

func TestNatsBus(t *testing.T) {
	// Start a NATS server for testing
	s, err := server.NewServer(&server.Options{Port: -1})
	assert.NoError(t, err)
	go s.Start()
	defer s.Shutdown()
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatalf("NATS server failed to start")
	}

	// Create a new NATS bus
	natsBusConfig := &bus.NatsBus{}
	natsBusConfig.SetServerUrl(s.ClientURL())
	conn, err := NewNatsConnection(natsBusConfig)
	assert.NoError(t, err)
	defer conn.Shutdown()
	bus := New[string](conn.GetClient())

	// Test Publish and Subscribe
	var receivedMsg string
	var mu sync.Mutex
	unsubscribe := bus.Subscribe(context.Background(), "test-topic", func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMsg = msg
	})
	defer unsubscribe()

	err = bus.Publish(context.Background(), "test-topic", "hello")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	assert.Equal(t, "hello", receivedMsg)
	mu.Unlock()

	// Test SubscribeOnce
	var receivedOnceMsg string
	unsubscribeOnce := bus.SubscribeOnce(context.Background(), "test-topic-once", func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedOnceMsg = msg
	})
	defer unsubscribeOnce()

	err = bus.Publish(context.Background(), "test-topic-once", "world")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	assert.Equal(t, "world", receivedOnceMsg)
	mu.Unlock()

	// Ensure the SubscribeOnce handler is not called again
	receivedOnceMsg = ""
	err = bus.Publish(context.Background(), "test-topic-once", "again")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, "", receivedOnceMsg)
}
