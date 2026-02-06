// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	busprotos "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/bus/redis"
	"github.com/stretchr/testify/require"
)

func TestRedisBus_ExternalServer(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	addr := mr.Addr()
	redisBusConfig := busprotos.RedisBus_builder{
		Address: &addr,
	}.Build()

	bus, err := redis.New[string](redisBusConfig)
	require.NoError(t, err)

	var receivedMsg string
	var mu sync.Mutex
	unsubscribe := bus.Subscribe(context.Background(), "test-topic", func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMsg = msg
	})
	defer unsubscribe()

	err = bus.Publish(context.Background(), "test-topic", "hello")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == "hello"
	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
}
