// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package kafka provides a Kafka implementation of the bus.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/proto/bus"
	kafkago "github.com/segmentio/kafka-go"
)

// writerInterface allows mocking kafka.Writer.
type writerInterface interface {
	WriteMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

// readerInterface allows mocking kafka.Reader.
type readerInterface interface {
	ReadMessage(ctx context.Context) (kafkago.Message, error)
	Close() error
}

// Bus is a Kafka-backed implementation of the Bus interface.
type Bus[T any] struct {
	writer        writerInterface
	brokers       []string
	topicPrefix   string
	consumerGroup string
	readerCreator func(config kafkago.ReaderConfig) readerInterface
}

// New creates and initializes a new KafkaBus.
//
// Parameters:
//   - config: *bus.KafkaBus. The configuration settings for the Kafka bus.
//
// Returns:
//   - *Bus[T]: A pointer to the initialized Kafka bus.
//   - error: An error if no brokers are provided or initialization fails.
func New[T any](config *bus.KafkaBus) (*Bus[T], error) {
	if len(config.GetBrokers()) == 0 {
		return nil, fmt.Errorf("kafka brokers are missing")
	}

	brokers := config.GetBrokers()
	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(brokers...),
		Balancer: &kafkago.LeastBytes{},
	}

	return &Bus[T]{
		writer:        writer,
		brokers:       brokers,
		topicPrefix:   config.GetTopicPrefix(),
		consumerGroup: config.GetConsumerGroup(),
		readerCreator: func(c kafkago.ReaderConfig) readerInterface {
			return kafkago.NewReader(c)
		},
	}, nil
}

// Publish sends a message to a Kafka topic.
//
// The message is marshaled to JSON and sent to the configured topic prefix + topic.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - topic: string. The topic to publish to.
//   - msg: T. The message payload.
//
// Returns:
//   - error: An error if marshaling or publishing fails.
func (b *Bus[T]) Publish(ctx context.Context, topic string, msg T) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	fullTopic := b.topicPrefix + topic

	err = b.writer.WriteMessages(ctx, kafkago.Message{
		Topic: fullTopic,
		Value: payload,
	})

	return err
}

// Subscribe subscribes to a Kafka topic.
//
// It starts a goroutine that continuously reads messages from the topic and invokes
// the provided handler.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function invoked for each message.
//
// Returns:
//   - func(): A function that unsubscribes the handler when called.
func (b *Bus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("kafka bus: handler cannot be nil")
		return func() {}
	}

	fullTopic := b.topicPrefix + topic

	groupID := b.consumerGroup
	if groupID == "" {
		// Broadcast behavior: unique group ID per instance ensures every instance gets the message.
		groupID = fmt.Sprintf("mcpany-%s", uuid.New().String())
	}

	readerConfig := kafkago.ReaderConfig{
		Brokers:  b.brokers,
		GroupID:  groupID,
		Topic:    fullTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}

	reader := b.readerCreator(readerConfig)

	// We use a derived context to handle cancellation from both parent context and unsubscribe
	ctx, cancel := context.WithCancel(ctx)
	var once sync.Once

	unsubscribe = func() {
		once.Do(func() {
			cancel()
			_ = reader.Close()
		})
	}

	go func() {
		defer unsubscribe()
		log := logging.GetLogger()

		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				// If context is done, we are shutting down or unsubscribed
				if ctx.Err() != nil {
					return
				}

				// Check for io.EOF or closed connection which might happen if Close() is called
				// explicitly while ReadMessage is blocking.
				// In kafka-go, Close() makes ReadMessage return error.
				return
			}

			var message T
			err = json.Unmarshal(m.Value, &message)
			if err != nil {
				log.Error("Failed to unmarshal message", "error", err)
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error("panic in handler", "error", r)
					}
				}()
				handler(message)
			}()
		}
	}()

	return unsubscribe
}

// SubscribeOnce subscribes to a topic for a single message.
//
// It ensures that the handler is called only once for the next message received.
//
// Parameters:
//   - ctx: context.Context. The context for the subscription.
//   - topic: string. The topic to subscribe to.
//   - handler: func(T). The callback function invoked for the single message.
//
// Returns:
//   - func(): A function that unsubscribes the handler if called before the message is received.
func (b *Bus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if handler == nil {
		logging.GetLogger().Error("kafka bus: handler cannot be nil")
		return func() {}
	}
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

// Close closes the Kafka writer.
//
// Summary: Closes the Kafka connection.
//
// Returns:
//   - error: An error if closing fails.
func (b *Bus[T]) Close() error {
	return b.writer.Close()
}
