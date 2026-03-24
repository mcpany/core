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

// Summary: Bus is a message bus implementation using NATS. Represents a Bus.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Bus[T any] struct {
	nc     *natsgo.Conn
	config *bus.NatsBus
	s      *server.Server
}

// Summary: New creates and initializes a new NATS bus. If the server URL is not provided in the configuration, an embedded NATS server is started on a random port.
//
// Parameters:
//   - config (*bus.NatsBus): The config parameter.
//
// Returns:
//   - *Bus[T]: The resulting *Bus[T].
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: Close closes the NATS bus connection and shuts down the embedded server if applicable. Closes the NATS connection.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (b *Bus[T]) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
	if b.s != nil {
		b.s.Shutdown()
	}
}

// Summary: Publish sends a message to a NATS topic. The message is marshaled to JSON before being published.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - topic (string): The topic parameter.
//   - msg (T): The msg parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (b *Bus[T]) Publish(_ context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.nc.Publish(topic, data)
}

// Summary: Subscribe registers a handler for a NATS topic. The handler will be invoked for each message received on the topic.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - topic (string): The topic parameter.
//   - handler (func(T)): The handler parameter.
//
// Returns:
//   - func(): The resulting func().
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SubscribeOnce registers a one-time handler for a NATS topic. The handler will be invoked only once for the next message received on the topic. The subscription is automatically removed after one message.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - topic (string): The topic parameter.
//   - handler (func(T)): The handler parameter.
//
// Returns:
//   - func(): The resulting func().
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
