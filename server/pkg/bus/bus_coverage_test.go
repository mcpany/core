// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"fmt"
	"testing"

	"github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
)

// TestGetBus_NatsError ...
// Summary: TestGetBus_NatsError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestGetBus_KafkaError ...
// Summary: TestGetBus_KafkaError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestGetBus_NatsSuccess ...
// Summary: TestGetBus_NatsSuccess
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestMessage_SetCorrelationID ...
// Summary: TestMessage_SetCorrelationID
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	msg := &BaseMessage{CID: "old"}
	msg.SetCorrelationID("new")
	assert.Equal(t, "new", msg.CorrelationID())
}

// TestNewProvider_Default ...
// Summary: TestNewProvider_Default
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test NewProvider with empty config (should default to InMemory)
	provider, err := NewProvider(nil)
	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// Verify it uses InMemory by checking type indirectly or just that it works
	bus, _ := GetBus[string](provider, "topic")
	assert.NotNil(t, bus)
}

// TestGetBus_Hook ...
// Summary: TestGetBus_Hook
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestGetBus_Hook_Success ...
// Summary: TestGetBus_Hook_Success
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestNewProvider_Hook ...
// Summary: TestNewProvider_Hook
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	oldHook := NewProviderHook
	defer func() { NewProviderHook = oldHook }()

	NewProviderHook = func(mb *bus.MessageBus) (*Provider, error) {
		return nil, fmt.Errorf("hook error")
	}

	_, err := NewProvider(nil)
	assert.Error(t, err)
	assert.Equal(t, "hook error", err.Error())
}
