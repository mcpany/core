// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	busprotos "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/bus/nats"
	"github.com/stretchr/testify/assert"
)

const testMessage = "hello"

// TestNatsBus_EmbeddedServer ...
// Summary: TestNatsBus_EmbeddedServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	err = bus.Publish(context.Background(), "test-topic", testMessage)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == testMessage
	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
}

// TestNatsBus_ExternalServer ...
// Summary: TestNatsBus_ExternalServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

	err = bus.Publish(context.Background(), "test-topic", testMessage)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == testMessage
	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
}
