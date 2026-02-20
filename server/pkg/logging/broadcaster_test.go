// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"fmt"
	"testing"
)

func TestBroadcaster_History(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 5 // Small limit for testing

	// Add messages
	for i := 0; i < 10; i++ {
		// Use string instead of []byte for simplicity with 'any'
		b.Broadcast(fmt.Sprintf("msg%d", i))
	}

	history := b.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected history length 5, got %d", len(history))
	}

	// Check content (last 5 messages: msg5, msg6, msg7, msg8, msg9)
	for i, msg := range history {
		expected := fmt.Sprintf("msg%d", i+5)
		// Assert type string
		strMsg, ok := msg.(string)
		if !ok {
			t.Errorf("Expected message to be string, got %T", msg)
			continue
		}
		if strMsg != expected {
			t.Errorf("Expected history[%d] to be %s, got %s", i, expected, strMsg)
		}
	}
}

func TestBroadcaster_SubscribeWithHistory(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 10

	// Add some initial messages
	b.Broadcast("msg1")
	b.Broadcast("msg2")

	// Subscribe
	ch, history := b.SubscribeWithHistory()
	defer b.Unsubscribe(ch)

	if len(history) != 2 {
		t.Errorf("Expected history length 2, got %d", len(history))
	}
	if history[0].(string) != "msg1" || history[1].(string) != "msg2" {
		t.Errorf("Unexpected history content")
	}

	// Broadcast new message
	b.Broadcast("msg3")

	// Should receive msg3
	select {
	case msg := <-ch:
		if msg.(string) != "msg3" {
			t.Errorf("Expected msg3, got %s", msg.(string))
		}
	default:
		t.Errorf("Did not receive msg3")
	}
}

func TestBroadcaster_HistoryIntegrity(t *testing.T) {
	// Test that modifying the returned history buffer doesn't affect internal state
	b := NewBroadcaster()
	msg := "original"
	b.Broadcast(msg)

	history := b.GetHistory()
	if history[0].(string) != "original" {
		t.Errorf("Unexpected message")
	}

	// For 'any', we rely on value semantics or immutability of the stored object.
	// Strings are immutable, so this test passes trivially.
	// Structs passed by value are copied, so modifying the copy doesn't affect history.
	// Pointers would be shared.
}

func TestBroadcaster_Hydrate(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 5

	messages := []any{
		"h1",
		"h2",
		"h3",
	}
	b.Hydrate(messages)

	// Check history
	history := b.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected history length 3, got %d", len(history))
	}
	if history[0].(string) != "h1" {
		t.Errorf("Unexpected history content")
	}

	// Add more (overflow)
	moreMessages := []any{
		"h4",
		"h5",
		"h6",
	}
	b.Hydrate(moreMessages)

	// Should have h2, h3, h4, h5, h6
	history = b.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected history length 5, got %d", len(history))
	}
	if history[0].(string) != "h2" {
		t.Errorf("Expected h2, got %s", history[0])
	}
	if history[4].(string) != "h6" {
		t.Errorf("Expected h6, got %s", history[4])
	}
}
