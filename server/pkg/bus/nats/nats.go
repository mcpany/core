// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package nats provides a NATS-based message bus implementation.
package nats

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mcpany/core/proto/bus"
	"github.com/nats-io/nats-server/v2/server"
	natsgo "github.com/nats-io/nats.go"
)

// Bus is a message bus implementation using NATS.
type Bus[T any] struct {
	nc     *natsgo.Conn
	config *bus.NatsBus
	s      *server.Server
}

// New creates a new NATS-backed Bus instance.
// If no server URL is provided in the config, an embedded NATS server is started.
//
// Parameters:
//   - config: *bus.NatsBus. The configuration for the NATS bus.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized NATS Bus.
//   - error: An error if connection fails or embedded server fails to start.
func New[T any](config *bus.NatsBus) (*Bus[T], error) {
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
	nc, err := natsgo.Connect(config.GetServerUrl())
	if err != nil {
		return nil, err
	}
	return &Bus[T]{
		nc:     nc,
		config: config,
		s:      s,
	}, nil
}

// Close closes the NATS connection and shuts down the embedded server if one was started.
//
// Returns:
//   None.
func (b *Bus[T]) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
	if b.s != nil {
		b.s.Shutdown()
	}
}

// Publish publishes a message to a NATS topic.
//
// Parameters:
//   - _: context.Context. Unused in NATS implementation (context is not propagated).
//   - topic: string. The topic/subject to publish to.
//   - msg: T. The message payload.
//
// Returns:
//   - error: An error if serialization or publishing fails.
func (b *Bus[T]) Publish(_ context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.nc.Publish(topic, data)
}

// Subscribe subscribes to a NATS topic and invokes the handler for each received message.
//
// Parameters:
//   - _: context.Context. Unused in NATS implementation.
//   - topic: string. The topic/subject to subscribe to.
//   - handler: func(T). The callback function.
//
// Returns:
//   - func(): A function to unsubscribe.
func (b *Bus[T]) Subscribe(_ context.Context, topic string, handler func(T)) (unsubscribe func()) {
	sub, _ := b.nc.Subscribe(topic, func(m *natsgo.Msg) {
		var msg T
		if err := json.Unmarshal(m.Data, &msg); err == nil {
			handler(msg)
		}
	})
	return func() {
		_ = sub.Unsubscribe()
	}
}

// SubscribeOnce subscribes to a NATS topic for a single message.
//
// Parameters:
//   - _: context.Context. Unused.
//   - topic: string. The topic/subject to subscribe to.
//   - handler: func(T). The callback function.
//
// Returns:
//   - func(): A function to unsubscribe (if called early).
func (b *Bus[T]) SubscribeOnce(_ context.Context, topic string, handler func(T)) (unsubscribe func()) {
	sub, err := b.nc.Subscribe(topic, func(m *natsgo.Msg) {
		var msg T
		if err := json.Unmarshal(m.Data, &msg); err == nil {
			handler(msg)
		}
	})
	if err != nil {
		return func() {}
	}
	_ = sub.AutoUnsubscribe(1)
	return func() {
		_ = sub.Unsubscribe()
	}
}
