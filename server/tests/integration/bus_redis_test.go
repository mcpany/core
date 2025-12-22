// Package integration tests the Redis bus implementation.

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisBus_Integration_Subscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	bus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic"
	msg := "hello"

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(m string) {
		assert.Equal(t, msg, m)
		wg.Done()
	}

	unsubscribe := bus.Subscribe(ctx, topic, handler)
	defer unsubscribe()

	// Give subscriber a moment to connect
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(ctx, topic, msg)
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_Integration_SubscribeOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode.")
	}

	redisAddr, cleanup := StartRedisContainer(t)
	defer cleanup()

	client := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	bus := redis.NewWithClient[string](client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	topic := "test-topic-once"
	msg := "hello once"

	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(m string) {
		assert.Equal(t, msg, m)
		wg.Done()
	}

	unsubscribe := bus.SubscribeOnce(ctx, topic, handler)
	defer unsubscribe()

	// Give subscriber a moment to connect
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(ctx, topic, msg)
	assert.NoError(t, err)

	wg.Wait()
}
