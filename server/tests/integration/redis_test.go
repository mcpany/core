// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	busprotos "github.com/mcpany/core/proto/bus"
	"github.com/mcpany/core/server/pkg/bus/redis"
	"github.com/stretchr/testify/require"
)

func TestRedisBus_ExternalServer(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TestRedisBus_ExternalServer in CI due to Docker requirement")
	}
	redisAddr, redisCleanup := StartRedisContainer(t)
	defer redisCleanup()

	redisBusConfig := busprotos.RedisBus_builder{
		Address: &redisAddr,
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
