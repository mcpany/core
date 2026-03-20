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
//
// Summary: Bus is a message bus implementation using NATS.
type Bus[T any] struct {
	nc     *natsgo.Conn
	config *bus.NatsBus
	s      *server.Server
}

// New creates and initializes a new NATS bus.
//
// Summary: New creates and initializes a new NATS bus.
//
// Parameters:
//   - config (*bus.NatsBus): The configuration settings.
//
// Returns:
//   - *Bus[T]: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// Close closes the NATS bus connection and shuts down the embedded server if applicable.
//
// Summary: Closes the NATS connection.
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

// Publish sends a message to a NATS topic.
//
// Summary: Publish sends a message to a NATS topic.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - topic (string): The textual representation of topic.
//   - msg (T): The provided msg data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (b *Bus[T]) Publish(_ context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.nc.Publish(topic, data)
}

// Subscribe registers a handler for a NATS topic.
//
// Summary: Subscribe registers a handler for a NATS topic.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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

// SubscribeOnce registers a one-time handler for a NATS topic.
//
// Summary: SubscribeOnce registers a one-time handler for a NATS topic.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - topic (string): The textual representation of topic.
//   - handler (func(T)): The provided handler data.
//
// Returns:
//   - unsubscribe (func()): The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
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
