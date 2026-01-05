// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

// import (
// 	"context"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/mcpany/core/server/pkg/bus/redis"
// 	busprotos "github.com/mcpany/core/proto/bus"
// 	"github.com/stretchr/testify/require"
// )

// func TestRedisBus_ExternalServer(t *testing.T) {
// 	if !IsDockerSocketAccessible() {
// 		t.Skip("Docker is not available, skipping test")
// 	}
// 	redisAddr, redisCleanup := StartRedisContainer(t)
// 	defer redisCleanup()

// 	redisBusConfig := &busprotos.RedisBus{Address: redisAddr}
// 	bus := redis.New[string](redisBusConfig)

// 	var receivedMsg string
// 	var mu sync.Mutex
// 	unsubscribe := bus.Subscribe(context.Background(), "test-topic", func(msg string) {
// 		mu.Lock()
// 		defer mu.Unlock()
// 		receivedMsg = msg
// 	})
// 	defer unsubscribe()

// 	err := bus.Publish(context.Background(), "test-topic", "hello")
// 	require.NoError(t, err)

// 	require.Eventually(t, func() bool {
// 		mu.Lock()
// 		defer mu.Unlock()
// 		return receivedMsg == "hello"
// 	}, 5*time.Second, 100*time.Millisecond, "did not receive message in time")
// }
