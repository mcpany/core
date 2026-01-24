// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"testing"
)

func TestBroadcaster_History(t *testing.T) {
	b := NewBroadcaster()
	// Set small limit for testing
	b.limit = 5

	msgs := [][]byte{
		[]byte("msg1"),
		[]byte("msg2"),
		[]byte("msg3"),
		[]byte("msg4"),
		[]byte("msg5"),
	}

	for _, m := range msgs {
		b.Broadcast(m)
	}

	// Verify history contains all 5
	hist := b.GetHistory()
	if len(hist) != 5 {
		t.Errorf("expected 5 history items, got %d", len(hist))
	}
	for i, m := range msgs {
		if !bytes.Equal(hist[i], m) {
			t.Errorf("history mismatch at index %d: expected %s, got %s", i, m, hist[i])
		}
	}

	// Add 6th message, should evict 1st
	msg6 := []byte("msg6")
	b.Broadcast(msg6)

	hist = b.GetHistory()
	if len(hist) != 5 {
		t.Errorf("expected 5 history items after overflow, got %d", len(hist))
	}

	expected := msgs[1:] // msg2, msg3, msg4, msg5
	expected = append(expected, msg6)

	for i, m := range expected {
		if !bytes.Equal(hist[i], m) {
			t.Errorf("history mismatch after overflow at index %d: expected %s, got %s", i, m, hist[i])
		}
	}
}

func TestBroadcaster_SubscribeWithHistory(t *testing.T) {
	b := NewBroadcaster()
	b.Broadcast([]byte("history1"))
	b.Broadcast([]byte("history2"))

	ch, hist := b.SubscribeWithHistory()
	defer b.Unsubscribe(ch)

	if len(hist) != 2 {
		t.Errorf("expected 2 history items, got %d", len(hist))
	}
	if string(hist[0]) != "history1" || string(hist[1]) != "history2" {
		t.Errorf("history mismatch")
	}

	// Send new message
	go func() {
		b.Broadcast([]byte("new1"))
	}()

	// Receive new message
	msg := <-ch
	if string(msg) != "new1" {
		t.Errorf("expected new message 'new1', got %s", msg)
	}
}
