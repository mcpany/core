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

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus/nats"
	busprotos "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

func TestNatsBus_EmbeddedServer(t *testing.T) {
	serverInfo := StartInProcessMCPANYServer(t, "embedded-nats")
	defer serverInfo.CleanupFunc()

	natsBusConfig := &busprotos.NatsBus{}
	bus, err := nats.New[string](natsBusConfig)
	assert.NoError(t, err)
	defer bus.Close()

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

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == "hello"
	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
}

func TestNatsBus_ExternalServer(t *testing.T) {
	serverInfo := StartMCPANYServer(t, "external-nats")
	defer serverInfo.CleanupFunc()

	natsBusConfig := &busprotos.NatsBus{}
	natsBusConfig.SetServerUrl(serverInfo.NatsURL)
	bus, err := nats.New[string](natsBusConfig)
	assert.NoError(t, err)
	defer bus.Close()

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

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == "hello"
	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
}
