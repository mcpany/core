
package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/proto/bus"
	"github.com/redis/go-redis/v9"
)

// RedisBus is a Redis-backed implementation of the Bus interface.
type RedisBus[T any] struct {
	client *redis.Client
	mu     sync.RWMutex
	// We need to keep track of pubsub clients to close them
	pubsubs map[string]*redis.PubSub
}

// New creates a new RedisBus.
func New[T any](redisConfig *bus.RedisBus) *RedisBus[T] {
	return NewWithClient[T](redis.NewClient(&redis.Options{
		Addr:     redisConfig.GetAddress(),
		Password: redisConfig.GetPassword(),
		DB:       int(redisConfig.GetDb()),
	}))
}

// NewWithClient creates a new RedisBus with an existing Redis client.
func NewWithClient[T any](client *redis.Client) *RedisBus[T] {
	return &RedisBus[T]{
		client:  client,
		pubsubs: make(map[string]*redis.PubSub),
	}
}

// Publish publishes a message to a Redis channel.
func (b *RedisBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, topic, payload).Err()
}

// Subscribe subscribes to a Redis channel.
func (b *RedisBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	pubsub := b.client.Subscribe(ctx, topic)
	b.pubsubs[topic] = pubsub

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var message T
			err := json.Unmarshal([]byte(msg.Payload), &message)
			if err != nil {
				log := logging.GetLogger()
				log.Error("Failed to unmarshal message", "error", err)
				continue
			}
			handler(message)
		}
	}()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if ps, ok := b.pubsubs[topic]; ok {
			ps.Close()
			delete(b.pubsubs, topic)
		}
	}
}

// SubscribeOnce subscribes to a topic for a single message.
func (b *RedisBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	var once sync.Once
	var unsub func()

	unsub = b.Subscribe(ctx, topic, func(msg T) {
		once.Do(func() {
			handler(msg)
			unsub()
		})
	})
	return unsub
}
