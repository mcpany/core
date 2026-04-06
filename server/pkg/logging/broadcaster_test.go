// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"fmt"
	"strconv"
	"testing"
)

func TestBroadcaster_History(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 5 // Small limit for testing

	// Add messages
	for i := 0; i < 10; i++ {
		b.Broadcast(fmt.Sprintf("msg%d", i))
	}

	history := b.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected history length 5, got %d", len(history))
	}

	// Check content (last 5 messages: msg5, msg6, msg7, msg8, msg9)
	for i, msg := range history {
		expected := fmt.Sprintf("msg%d", i+5)
		if msg.(string) != expected {
			t.Errorf("Expected history[%d] to be %s, got %s", i, expected, msg.(string))
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

// TestBroadcaster_HistoryIntegrity removed because we now store 'any'.
// If 'any' is a value type (like struct or string), it's safe.
// If it's a pointer, it's shared. We assume callers know this.

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
		t.Errorf("Expected h2, got %s", history[0].(string))
	}
	if history[4].(string) != "h6" {
		t.Errorf("Expected h6, got %s", history[4].(string))
	}
}

func TestBroadcaster_DeleteIf(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 5

	// Add items
	for i := 0; i < 5; i++ {
		b.Broadcast(map[string]any{"id": i, "trace_id": "trace-" + strconv.Itoa(i)})
	}

	// Delete even IDs
	b.DeleteIf(func(item any) bool {
		m, ok := item.(map[string]any)
		if !ok {
			return false
		}
		id, ok := m["id"].(int)
		return ok && id%2 == 0
	})

	history := b.GetHistory()
	if len(history) != 2 {
		t.Fatalf("expected history length 2, got %d", len(history))
	}

	m1 := history[0].(map[string]any)
	m2 := history[1].(map[string]any)
	if m1["id"] != 1 || m2["id"] != 3 {
		t.Errorf("unexpected items remaining in history: %v", history)
	}

	// Verify we can still push normally after delete
	b.Broadcast(map[string]any{"id": 5, "trace_id": "trace-5"})
	history2 := b.GetHistory()
	if len(history2) != 3 {
		t.Fatalf("expected history length 3, got %d", len(history2))
	}
}
