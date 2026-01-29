package bus

import (
	"fmt"
	"testing"

	"github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

func TestGetBus_NatsError(t *testing.T) {
	// Configure NATS with invalid URL to trigger error in GetBus
	msgBus := bus.MessageBus_builder{}.Build()
	natsBus := bus.NatsBus_builder{}.Build()
	natsBus.SetServerUrl("nats://invalid:1234")
	msgBus.SetNats(natsBus)

	provider, err := NewProvider(msgBus)
	assert.NoError(t, err)

	_, err = GetBus[string](provider, "test-topic")
	assert.Error(t, err)
}

func TestGetBus_KafkaError(t *testing.T) {
	msgBus := bus.MessageBus_builder{}.Build()
	kafkaBus := bus.KafkaBus_builder{}.Build()
	// No brokers set
	msgBus.SetKafka(kafkaBus)

	provider, err := NewProvider(msgBus)
	assert.NoError(t, err)

	_, err = GetBus[string](provider, "test-topic")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka brokers are missing")
}

func TestGetBus_NatsSuccess(t *testing.T) {
	// Configure NATS with empty URL to trigger embedded server
	msgBus := bus.MessageBus_builder{}.Build()
	natsBus := bus.NatsBus_builder{}.Build()
	msgBus.SetNats(natsBus)

	provider, err := NewProvider(msgBus)
	assert.NoError(t, err)

	// This should start embedded NATS and succeed
	bus, err := GetBus[string](provider, "test-topic")
	assert.NoError(t, err)
	assert.NotNil(t, bus)
}

func TestMessage_SetCorrelationID(t *testing.T) {
	msg := &BaseMessage{CID: "old"}
	msg.SetCorrelationID("new")
	assert.Equal(t, "new", msg.CorrelationID())
}

func TestNewProvider_Default(t *testing.T) {
	// Test NewProvider with empty config (should default to InMemory)
	provider, err := NewProvider(nil)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Verify it uses InMemory by checking type indirectly or just that it works
	bus, _ := GetBus[string](provider, "topic")
	assert.NotNil(t, bus)
}

func TestGetBus_Hook(t *testing.T) {
	oldHook := GetBusHook
	defer func() { GetBusHook = oldHook }()

	GetBusHook = func(p *Provider, topic string) (any, error) {
		return nil, fmt.Errorf("hook error")
	}

	provider, _ := NewProvider(nil)
	_, err := GetBus[string](provider, "topic")
	assert.Error(t, err)
	assert.Equal(t, "hook error", err.Error())
}

func TestGetBus_Hook_Success(t *testing.T) {
	oldHook := GetBusHook
	defer func() { GetBusHook = oldHook }()

	mockBus := &MockBus[string]{}
	GetBusHook = func(p *Provider, topic string) (any, error) {
		return mockBus, nil
	}

	provider, _ := NewProvider(nil)
	bus, err := GetBus[string](provider, "topic")
	assert.NoError(t, err)
	assert.Equal(t, mockBus, bus)
}

func TestNewProvider_Hook(t *testing.T) {
	oldHook := NewProviderHook
	defer func() { NewProviderHook = oldHook }()

	NewProviderHook = func(mb *bus.MessageBus) (*Provider, error) {
		return nil, fmt.Errorf("hook error")
	}

	_, err := NewProvider(nil)
	assert.Error(t, err)
	assert.Equal(t, "hook error", err.Error())
}
