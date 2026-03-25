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
// Summary: Represents a Bus.
type Bus[T any] struct {
	nc     *natsgo.Conn
	config *bus.NatsBus
	s      *server.Server
}

// New creates and initializes a new NATS bus.
//
// If the server URL is not provided in the configuration, an embedded NATS server
// is started on a random port.
//
// Parameters: - None.
//   - config: *bus.NatsBus. The configuration settings for the NATS bus.
//
// Returns: - None.
//   - *Bus[T]: A pointer to the initialized NATS bus.
//   - error: An error if the connection or embedded server startup fails.
//
// Summary: Initializes New operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
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

// Close closes the NATS bus connection and shuts down the embedded server if applicable.
//
// Summary: Closes the NATS connection.
//
// Returns: - None.
//
//	None.
func (b *Bus[T]) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
	if b.s != nil {
		b.s.Shutdown()
	}
}

// Publish publish publish.
//
// Summary: Publish publish.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - topic (string): The topic.
//   - msg (T): The msg.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (b *Bus[T]) Publish(_ context.Context, topic string, msg T) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return b.nc.Publish(topic, data)
}

// Subscribe subscribe subscribe.
//
// Summary: Subscribe subscribe.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - topic (string): The topic.
//   - handler (func(T)): The handler.
//
// Returns: - None.
//   - func(): The result.
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

// SubscribeOnce subscribeOnce subscribe once.
//
// Summary: SubscribeOnce subscribe once.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - topic (string): The topic.
//   - handler (func(T)): The handler.
//
// Returns: - None.
//   - func(): The result.
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
