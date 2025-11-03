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

package nats

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats.go"
)

// NatsBus is a message bus implementation using NATS.
type NatsBus[T any] struct {
	nc     *nats.Conn
	config *bus.NatsBus
	mu     sync.Mutex
}

// New creates a new NATS bus.
func New[T any](config *bus.NatsBus) (*NatsBus[T], error) {
	nc, err := nats.Connect(config.GetServerUrl())
	if err != nil {
		return nil, err
	}
	return &NatsBus[T]{
		nc:     nc,
		config: config,
	}, nil
}

// Publish sends a message to a NATS topic.
func (b *NatsBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.nc.Publish(topic, data)
}

// Subscribe registers a handler for a NATS topic.
func (b *NatsBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	sub, _ := b.nc.Subscribe(topic, func(m *nats.Msg) {
		var msg T
		if err := json.Unmarshal(m.Data, &msg); err == nil {
			handler(msg)
		}
	})
	return func() {
		sub.Unsubscribe()
	}
}

// SubscribeOnce registers a one-time handler for a NATS topic.
func (b *NatsBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	sub, err := b.nc.Subscribe(topic, func(m *nats.Msg) {
		var msg T
		if err := json.Unmarshal(m.Data, &msg); err == nil {
			handler(msg)
		}
	})
	if err != nil {
		return func() {}
	}
	sub.AutoUnsubscribe(1)
	return func() {
		sub.Unsubscribe()
	}
}
