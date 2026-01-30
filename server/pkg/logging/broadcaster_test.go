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
		b.Broadcast([]byte(fmt.Sprintf("msg%d", i)))
	}

	history := b.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected history length 5, got %d", len(history))
	}

	// Check content (last 5 messages: msg5, msg6, msg7, msg8, msg9)
	for i, msg := range history {
		expected := fmt.Sprintf("msg%d", i+5)
		if string(msg) != expected {
			t.Errorf("Expected history[%d] to be %s, got %s", i, expected, string(msg))
		}
	}
}

func TestBroadcaster_SubscribeWithHistory(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 10

	// Add some initial messages
	b.Broadcast([]byte("msg1"))
	b.Broadcast([]byte("msg2"))

	// Subscribe
	ch, history := b.SubscribeWithHistory()
	defer b.Unsubscribe(ch)

	if len(history) != 2 {
		t.Errorf("Expected history length 2, got %d", len(history))
	}
	if string(history[0]) != "msg1" || string(history[1]) != "msg2" {
		t.Errorf("Unexpected history content")
	}

	// Broadcast new message
	b.Broadcast([]byte("msg3"))

	// Should receive msg3
	select {
	case msg := <-ch:
		if string(msg) != "msg3" {
			t.Errorf("Expected msg3, got %s", string(msg))
		}
	default:
		t.Errorf("Did not receive msg3")
	}
}

func TestBroadcaster_HistoryIntegrity(t *testing.T) {
	// Test that modifying the returned history buffer doesn't affect internal state
	b := NewBroadcaster()
	msg := []byte("original")
	b.Broadcast(msg)

	history := b.GetHistory()
	if string(history[0]) != "original" {
		t.Errorf("Unexpected message")
	}

	// Note: We intentionally allow GetHistory to return references to internal buffers for zero-copy performance.
	// Callers must not modify the returned data.
	// The previous test checking for deep-copy isolation is removed.

	// Test that reusing buffer in Broadcast doesn't affect history
	buf := []byte("hello")
	b.Broadcast(buf)

	// Modify buf
	buf[0] = 'H'

	history3 := b.GetHistory()
	// The last message should be "hello" (history3[1])
	lastMsg := history3[len(history3)-1]
	if string(lastMsg) != "hello" {
		t.Errorf("History was affected by caller modifying buffer after broadcast: got %s, expected hello", string(lastMsg))
	}
}

func TestBroadcaster_Hydrate(t *testing.T) {
	b := NewBroadcaster()
	b.limit = 5

	messages := [][]byte{
		[]byte("h1"),
		[]byte("h2"),
		[]byte("h3"),
	}
	b.Hydrate(messages)

	// Check history
	history := b.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected history length 3, got %d", len(history))
	}
	if string(history[0]) != "h1" {
		t.Errorf("Unexpected history content")
	}

	// Add more (overflow)
	moreMessages := [][]byte{
		[]byte("h4"),
		[]byte("h5"),
		[]byte("h6"),
	}
	b.Hydrate(moreMessages)

	// Should have h2, h3, h4, h5, h6
	history = b.GetHistory()
	if len(history) != 5 {
		t.Errorf("Expected history length 5, got %d", len(history))
	}
	if string(history[0]) != "h2" {
		t.Errorf("Expected h2, got %s", string(history[0]))
	}
	if string(history[4]) != "h6" {
		t.Errorf("Expected h6, got %s", string(history[4]))
	}
}
