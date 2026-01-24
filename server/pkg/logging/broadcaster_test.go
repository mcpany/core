// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroadcasterHistory(t *testing.T) {
	b := NewBroadcaster()
	// Lower limit for testing
	b.limit = 5

	// Add messages
	msgs := [][]byte{
		[]byte("msg1"),
		[]byte("msg2"),
		[]byte("msg3"),
	}

	for _, msg := range msgs {
		b.Broadcast(msg)
	}

	// Verify history contains all
	history := b.GetHistory()
	assert.Len(t, history, 3)
	assert.Equal(t, msgs, history)

	// Add more to overflow limit
	moreMsgs := [][]byte{
		[]byte("msg4"),
		[]byte("msg5"),
		[]byte("msg6"),
	}
	for _, msg := range moreMsgs {
		b.Broadcast(msg)
	}

	// Should have 5 messages: msg2, msg3, msg4, msg5, msg6
	history = b.GetHistory()
	assert.Len(t, history, 5)
	assert.Equal(t, []byte("msg2"), history[0])
	assert.Equal(t, []byte("msg6"), history[4])
}

func TestSubscribeWithHistory(t *testing.T) {
	b := NewBroadcaster()
	msg1 := []byte("history1")
	b.Broadcast(msg1)

	// Subscribe
	ch, history := b.SubscribeWithHistory()

	// Verify history received
	assert.Len(t, history, 1)
	assert.Equal(t, msg1, history[0])

	// Verify channel works for new messages
	msg2 := []byte("new1")
	b.Broadcast(msg2)

	select {
	case received := <-ch:
		assert.Equal(t, msg2, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch")
	}

	b.Unsubscribe(ch)
}
