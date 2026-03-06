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

// Bus - Auto-generated documentation.
//
// Summary: Bus is a message bus implementation using NATS.
//
// Fields:
//   - Various fields for Bus.
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
// Parameters:
//   - config: *bus.NatsBus. The configuration settings for the NATS bus.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized NATS bus.
//   - error: An error if the connection or embedded server startup fails.
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

// Close - Auto-generated documentation.
//
// Summary: Close closes the NATS bus connection and shuts down the embedded server if applicable.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func (b *Bus[T]) Close() {
	if b.nc != nil {
		b.nc.Close()
	}
	if b.s != nil {
		b.s.Shutdown()
	}
}

// Publish sends a message to a NATS topic. The message is marshaled to JSON before being published. Parameters: - _: context.Context. The context (unused in NATS publish). - topic: string. The topic to publish to. - msg: T. The message payload. Returns: - error: An error if marshaling or publishing fails.
//
// Summary: Publish sends a message to a NATS topic. The message is marshaled to JSON before being published. Parameters: - _: context.Context. The context (unused in NATS publish). - topic: string. The topic to publish to. - msg: T. The message payload. Returns: - error: An error if marshaling or publishing fails.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - topic (string): The topic parameter used in the operation.
//   - msg (T): The msg parameter used in the operation.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
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

// Subscribe registers a handler for a NATS topic. The handler will be invoked for each message received on the topic. Parameters: - _: context.Context. The context (unused in NATS subscribe). - topic: string. The topic to subscribe to. - handler: func(T). The callback function invoked for each message. Returns: - func(): A function that unsubscribes the handler when called.
//
// Summary: Subscribe registers a handler for a NATS topic. The handler will be invoked for each message received on the topic. Parameters: - _: context.Context. The context (unused in NATS subscribe). - topic: string. The topic to subscribe to. - handler: func(T). The callback function invoked for each message. Returns: - func(): A function that unsubscribes the handler when called.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - topic (string): The topic parameter used in the operation.
//   - handler (func(T): The handler parameter used in the operation.
//
// Returns:
//   - None.
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

// SubscribeOnce registers a one-time handler for a NATS topic. The handler will be invoked only once for the next message received on the topic. The subscription is automatically removed after one message. Parameters: - _: context.Context. The context (unused in NATS subscribe). - topic: string. The topic to subscribe to. - handler: func(T). The callback function invoked for the single message. Returns: - func(): A function that unsubscribes the handler if called before the message is received.
//
// Summary: SubscribeOnce registers a one-time handler for a NATS topic. The handler will be invoked only once for the next message received on the topic. The subscription is automatically removed after one message. Parameters: - _: context.Context. The context (unused in NATS subscribe). - topic: string. The topic to subscribe to. - handler: func(T). The callback function invoked for the single message. Returns: - func(): A function that unsubscribes the handler if called before the message is received.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - topic (string): The topic parameter used in the operation.
//   - handler (func(T): The handler parameter used in the operation.
//
// Returns:
//   - None.
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
