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

	"github.com/mcpany/core/pkg/bus/redis"
	"github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRedisBus_SubscribeOnce(t *testing.T) {
	if !IsDockerSocketAccessible() {
		t.Skip("Docker is not available, skipping test")
	}
	redisAddr, redisCleanup := StartRedisContainer(t)
	defer redisCleanup()

	redisBusConfig := bus.RedisBus_builder{
		Address:  &redisAddr,
		Db:       proto.Int32(0),
		Password: proto.String(""),
	}.Build()
	bus := redis.New[string](redisBusConfig)

	topic := "test-topic"
	expectedMsg := "hello"
	var receivedMsg string
	var mu sync.Mutex

	unsubscribe := bus.SubscribeOnce(context.Background(), topic, func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		receivedMsg = msg
	})
	defer unsubscribe()

	err := bus.Publish(context.Background(), topic, expectedMsg)
	assert.NoError(t, err)

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return receivedMsg == expectedMsg
	}, time.Second, 10*time.Millisecond, "did not receive message in time")
}
