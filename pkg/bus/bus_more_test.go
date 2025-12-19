// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"testing"
	"time"

	buspb "github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

func TestNatsBus_Integration(t *testing.T) {
	// Configure NATS bus with empty server URL to trigger embedded server
	natsConfig := buspb.NatsBus_builder{}.Build()
	messageBus := buspb.MessageBus_builder{}.Build()
	messageBus.SetNats(natsConfig)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	topic := "test-nats-topic"
	bus := GetBus[string](provider, topic)
	assert.NotNil(t, bus)

	done := make(chan string, 1)
	bus.Subscribe(context.Background(), topic, func(msg string) {
		done <- msg
	})

	// NATS subscription is synchronous for the client, so we should be ready to publish.
	err = bus.Publish(context.Background(), topic, "hello nats")
	assert.NoError(t, err)

	select {
	case msg := <-done:
		assert.Equal(t, "hello nats", msg)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for nats message")
	}
}

func TestKafkaBus_Initialization(t *testing.T) {
	// Configure Kafka bus with dummy brokers
	kafkaConfig := buspb.KafkaBus_builder{}.Build()
	kafkaConfig.SetBrokers([]string{"localhost:9092"})
	kafkaConfig.SetTopicPrefix("test-")

	messageBus := buspb.MessageBus_builder{}.Build()
	messageBus.SetKafka(kafkaConfig)

	provider, err := NewProvider(messageBus)
	assert.NoError(t, err)

	topic := "test-kafka-topic"
	bus := GetBus[string](provider, topic)
	assert.NotNil(t, bus)

	// We cannot easily test Publish/Subscribe without a real Kafka broker
	// but we verified that GetBus creates the bus instance correctly.
}
