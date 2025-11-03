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
	"time"

	"github.com/mcpany/core/proto/bus"
	s "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// NatsConnection manages the connection to a NATS server.
type NatsConnection struct {
	nc *nats.Conn
	ns *s.Server
}

// NewNatsConnection creates a new NATS connection.
func NewNatsConnection(config *bus.NatsBus) (*NatsConnection, error) {
	serverURL := config.GetServerUrl()
	var ns *s.Server
	if serverURL == "" {
		var err error
		ns, err = s.NewServer(&s.Options{Port: -1})
		if err != nil {
			return nil, err
		}
		go ns.Start()
		if !ns.ReadyForConnections(4 * time.Second) {
			return nil, err
		}
		serverURL = ns.ClientURL()
	}

	nc, err := nats.Connect(serverURL)
	if err != nil {
		return nil, err
	}
	return &NatsConnection{
		nc: nc,
		ns: ns,
	}, nil
}

// Shutdown shuts down the NATS server if it was started by the bus.
func (c *NatsConnection) Shutdown() {
	if c.ns != nil {
		c.ns.Shutdown()
	}
}

// GetClient returns the NATS client.
func (c *NatsConnection) GetClient() *nats.Conn {
	return c.nc
}

// NatsBus is a message bus implementation using NATS.
type NatsBus[T any] struct {
	nc *nats.Conn
	mu sync.Mutex
}

// New creates a new NATS bus.
func New[T any](nc *nats.Conn) *NatsBus[T] {
	return &NatsBus[T]{
		nc: nc,
	}
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
