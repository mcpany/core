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

package main

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	buspb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

func TestSetup_NoErrorOnValidRedis(t *testing.T) {
	// Set a valid REDIS_ADDR
	redisAddr := "redis://localhost:6379"
	os.Setenv("REDIS_ADDR", redisAddr)
	defer os.Unsetenv("REDIS_ADDR")

	// Skip test if Redis is not available
	opts, err := redis.ParseURL(redisAddr)
	if err != nil {
		t.Fatalf("invalid redis url: %v", err)
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	_, err = setup()
	assert.NoError(t, err, "setup() should not return an error with a valid REDIS_ADDR")
}

func TestSetup_InMemoryBusWhenRedisAddrNotSet(t *testing.T) {
	// Unset REDIS_ADDR to ensure in-memory bus is used
	os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	assert.NoError(t, err, "setup() should not return an error when REDIS_ADDR is not set")
}

func TestSetup_ErrorOnInvalidRedisAddr(t *testing.T) {
	// Set an invalid REDIS_ADDR
	os.Setenv("REDIS_ADDR", "invalid-address")
	defer os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	assert.Error(t, err, "setup() should return an error with an invalid REDIS_ADDR")
}

func TestSetup_InMemoryBus(t *testing.T) {
	// Unset REDIS_ADDR to fall back to in-memory bus
	os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	assert.NoError(t, err, "setup() should not return an error when using in-memory bus")
}

func TestSetup_DirectValidationError(t *testing.T) {
	// This test directly manipulates the config to trigger an error in the validation logic
	// to ensure the error handling in setup() is covered.
	busConfig := &buspb.MessageBus{}
	busConfig.SetRedis(&buspb.RedisBus{}) // Set an empty redis bus to trigger a validation error

	_, err := setupWithConfig(busConfig)
	assert.Error(t, err, "setupWithConfig() should return an error with an invalid busConfig")
}
