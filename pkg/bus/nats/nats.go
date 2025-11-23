// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// NatsBus is a message bus implementation using NATS.
type NatsBus[T any] struct {
	nc     *nats.Conn
	config *bus.NatsBus
	s      *server.Server
	mu     sync.Mutex
}

// New creates a new NATS bus.
func New[T any](config *bus.NatsBus) (*NatsBus[T], error) {
	var s *server.Server
	if config.GetServerUrl() == "" {
		var err error
		s, err = server.NewServer(&server.Options{Port: -1})
		if err != nil {
			return nil, err
		}
		go s.Start()
		if !s.ReadyForConnections(4 * time.Second) {
			s.Shutdown()
			return nil, errors.New("nats server failed to start")
		}
		config.SetServerUrl(s.ClientURL())
	}
	nc, err := nats.Connect(config.GetServerUrl())
	if err != nil {
		return nil, err
	}
	return &NatsBus[T]{
		nc:     nc,
		config: config,
		s:      s,
	}, nil
}

// Close closes the NATS bus.
func (b *NatsBus[T]) Close() {
	if b.s != nil {
		b.s.Shutdown()
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
