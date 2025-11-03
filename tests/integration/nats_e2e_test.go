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

	"github.com/mcpany/core/pkg/bus"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNatsBusEndToEnd(t *testing.T) {
	// Server setup
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetNats(bus_pb.NatsBus_builder{}.Build())
	busProvider, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)
	defer busProvider.Close()

	natsBus := bus.GetBus[string](busProvider, "test-topic")

	// Test Publish and Subscribe
	var receivedMsg string
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	unsubscribe := natsBus.Subscribe(context.Background(), "test-topic", func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMsg = msg
		wg.Done()
	})
	defer unsubscribe()

	err = natsBus.Publish(context.Background(), "test-topic", "hello")
	assert.NoError(t, err)

	wg.Wait()
	mu.Lock()
	assert.Equal(t, "hello", receivedMsg)
	mu.Unlock()
}
