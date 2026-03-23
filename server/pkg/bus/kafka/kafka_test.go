// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/mcpany/core/proto/bus"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWriter is a mock for kafkaWriter
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) WriteMessages(ctx context.Context, msgs ...kafkago.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockReader is a mock for kafkaReader
type MockReader struct {
	mock.Mock
}

func (m *MockReader) ReadMessage(ctx context.Context) (kafkago.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafkago.Message), args.Error(1)
}

func (m *MockReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestPublish(t *testing.T) {
	mockWriter := new(MockWriter)
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"127.0.0.1:9092"})

	b, err := New[string](config)
	assert.NoError(t, err)

	b.writer = mockWriter // Inject mock

	ctx := context.Background()
	msg := "test-message"
	payload, _ := json.Marshal(msg)

	mockWriter.On("WriteMessages", ctx, []kafkago.Message{{
		Topic: "test-topic",
		Value: payload,
	}}).Return(nil)

	err = b.Publish(ctx, "test-topic", msg)
	assert.NoError(t, err)
	mockWriter.AssertExpectations(t)
}

func TestSubscribe(t *testing.T) {
	mockWriter := new(MockWriter) // Not used but needed for New
	mockReader := new(MockReader)

	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"127.0.0.1:9092"})

	b, err := New[string](config)
	assert.NoError(t, err)

	// Inject mocks
	b.writer = mockWriter
	b.readerCreator = func(c kafkago.ReaderConfig) readerInterface {
		assert.Equal(t, "test-topic", c.Topic)
		return mockReader
	}

	ctx := context.Background()
	msg := "test-message"
	payload, _ := json.Marshal(msg)

	// Setup mock reader behavior
	// First call returns message
	mockReader.On("ReadMessage", mock.Anything).Return(kafkago.Message{
		Value: payload,
	}, nil).Once()

	// Second call returns error to stop loop
	mockReader.On("ReadMessage", mock.Anything).Return(kafkago.Message{}, fmt.Errorf("stop")).Once()

	mockReader.On("Close").Return(nil)

	received := make(chan string)
	unsubscribe := b.Subscribe(ctx, "test-topic", func(m string) {
		received <- m
	})

	select {
	case m := <-received:
		assert.Equal(t, msg, m)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	unsubscribe()
	// Allow goroutine to finish
	time.Sleep(100 * time.Millisecond)
}

func TestSubscribeOnce(t *testing.T) {
	mockWriter := new(MockWriter)
	mockReader := new(MockReader)

	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"127.0.0.1:9092"})

	b, err := New[string](config)
	assert.NoError(t, err)

	b.writer = mockWriter
	b.readerCreator = func(c kafkago.ReaderConfig) readerInterface {
		return mockReader
	}

	ctx := context.Background()
	msg := "test-message"
	payload, _ := json.Marshal(msg)

	// ReadMessage called once
	mockReader.On("ReadMessage", mock.Anything).Return(kafkago.Message{
		Value: payload,
	}, nil).Once()

	// It might be called again before unsubscribe happens in the handler race, or unsubscribe handles closing.
	// When SubscribeOnce handler is called, it calls unsubscribe.
	// unsubscribe calls reader.Close().
	// ReadMessage might return error then.

	mockReader.On("ReadMessage", mock.Anything).Return(kafkago.Message{}, fmt.Errorf("closed")).Maybe()

	mockReader.On("Close").Return(nil)

	received := make(chan string)
	b.SubscribeOnce(ctx, "test-topic", func(m string) {
		received <- m
	})

	select {
	case m := <-received:
		assert.Equal(t, msg, m)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	time.Sleep(100 * time.Millisecond)
}

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	config.SetBrokers([]string{"127.0.0.1:9092"})

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
	config.SetBrokers([]string{"127.0.0.1:9092"})

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
	config.SetBrokers([]string{"127.0.0.1:9092"})
	b, err := New[string](config)
	assert.NoError(t, err)

	unsubscribe := b.Subscribe(context.Background(), "topic", nil)
	assert.NotNil(t, unsubscribe)
	unsubscribe()
}

func TestSubscribeOnce_HandlerNil(t *testing.T) {
	config := &bus.KafkaBus{}
	config.SetBrokers([]string{"127.0.0.1:9092"})
	b, err := New[string](config)
	assert.NoError(t, err)

	unsubscribe := b.SubscribeOnce(context.Background(), "topic", nil)
	assert.NotNil(t, unsubscribe)
	unsubscribe()
}
