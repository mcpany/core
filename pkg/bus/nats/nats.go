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
	"fmt"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// ConnectionManager manages a single NATS connection and an optional embedded server.
type ConnectionManager struct {
	Nc     *nats.Conn
	Server *server.Server
}

// NewConnectionManager creates a new NATS connection manager.
func NewConnectionManager(config *bus.NatsBus) (*ConnectionManager, error) {
	var s *server.Server
	serverURL := config.GetServerUrl()
	if serverURL == "" {
		var err error
		s, err = server.NewServer(&server.Options{Port: -1})
		if err != nil {
			return nil, err
		}
		go s.Start()
		if !s.ReadyForConnections(4 * time.Second) {
			s.Shutdown()
			return nil, fmt.Errorf("embedded NATS server failed to start")
		}
		serverURL = s.ClientURL()
	}
	nc, err := nats.Connect(serverURL)
	if err != nil {
		if s != nil {
			s.Shutdown()
		}
		return nil, err
	}
	return &ConnectionManager{
		Nc:     nc,
		Server: s,
	}, nil
}

// Close closes the NATS connection and shuts down the embedded server if it exists.
func (cm *ConnectionManager) Close() error {
	cm.Nc.Close()
	if cm.Server != nil {
		cm.Server.Shutdown()
	}
	return nil
}

// NatsBus is a message bus implementation using NATS.
type NatsBus[T any] struct {
	cm *ConnectionManager
}

// New creates a new NATS bus.
func New[T any](cm *ConnectionManager) *NatsBus[T] {
	return &NatsBus[T]{
		cm: cm,
	}
}

// Publish sends a message to a NATS topic.
func (b *NatsBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.cm.Nc.Publish(topic, data)
}

// Subscribe registers a handler for a NATS topic.
func (b *NatsBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	sub, _ := b.cm.Nc.Subscribe(topic, func(m *nats.Msg) {
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
	sub, err := b.cm.Nc.Subscribe(topic, func(m *nats.Msg) {
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
