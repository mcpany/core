// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/proto/bus"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestNew_Errors(t *testing.T) {
	// Missing brokers
	config := &bus.KafkaBus{}
	b, err := New[string](config)
	assert.Error(t, err)
	assert.Nil(t, b)
	assert.Contains(t, err.Error(), "kafka brokers are missing")
}

func TestClose(t *testing.T) {
	mockWriter := new(MockWriter)
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"localhost:9092"})

	b, err := New[string](config)
	assert.NoError(t, err)

	b.writer = mockWriter // Inject mock
	mockWriter.On("Close").Return(nil)

	err = b.Close()
	assert.NoError(t, err)
	mockWriter.AssertExpectations(t)
}

func TestPublish_Error(t *testing.T) {
	mockWriter := new(MockWriter)
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"localhost:9092"})

	b, err := New[string](config)
	assert.NoError(t, err)

	b.writer = mockWriter // Inject mock

	ctx := context.Background()
	msg := "test-message"
	payload, _ := json.Marshal(msg)

	mockWriter.On("WriteMessages", ctx, []kafkago.Message{{
		Topic: "test-topic",
		Value: payload,
	}}).Return(assert.AnError)

	err = b.Publish(ctx, "test-topic", msg)
	assert.Error(t, err)
	mockWriter.AssertExpectations(t)
}

func TestSubscribe_HandlerNil(t *testing.T) {
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"localhost:9092"})
	b, err := New[string](config)
	assert.NoError(t, err)

	unsubscribe := b.Subscribe(context.Background(), "topic", nil)
	assert.NotNil(t, unsubscribe)
	unsubscribe()
}

func TestSubscribeOnce_HandlerNil(t *testing.T) {
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"localhost:9092"})
	b, err := New[string](config)
	assert.NoError(t, err)

	unsubscribe := b.SubscribeOnce(context.Background(), "topic", nil)
	assert.NotNil(t, unsubscribe)
	unsubscribe()
}
