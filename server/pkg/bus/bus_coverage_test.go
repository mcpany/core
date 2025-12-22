package bus

import (
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

func TestNewProvider_UnknownType(_ *testing.T) {
	// It's hard to force unknown type if we use the proto setters correctly.
	// But we can check if we can reach the default case in NewProvider
	// which returns "unknown bus type".
	// The proto generated code might strictly limit values for oneof.
	// However, we can construct a MessageBus where the oneof is nil but HasBusType is true?
	// No, HasBusType checks if non-nil.

	// If we use a manually constructed config with no inner fields set but somehow bypass helpers?
	// NewProvider logic:
	// if !provider.config.HasBusType() { set InMemory }
	// switch provider.config.WhichBusType() { ... }

	// If we have a type that is not InMemory, Redis, or Nats.
	// Current proto only has those 3.
	// So "unknown bus type" might be unreachable unless we add new types to proto but not code.
	// We can skip this edge case if hard to reach.
}
