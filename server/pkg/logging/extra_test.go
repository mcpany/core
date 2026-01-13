// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	b := NewBroadcaster()

	// Test Subscribe
	ch1 := b.Subscribe()
	ch2 := b.Subscribe()
	assert.Len(t, b.subscribers, 2)

	// Test Broadcast
	msg := []byte("test message")
	b.Broadcast(msg)

	select {
	case received := <-ch1:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch1")
	}

	select {
	case received := <-ch2:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch2")
	}

	// Test Unsubscribe
	b.Unsubscribe(ch1)
	assert.Len(t, b.subscribers, 1)

	// Ensure ch1 is closed
	_, ok := <-ch1
	assert.False(t, ok)

	// Broadcast again, only ch2 should receive
	msg2 := []byte("test message 2")
	b.Broadcast(msg2)

	select {
	case received := <-ch2:
		assert.Equal(t, msg2, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for ch2")
	}

	// Unsubscribe ch2
	b.Unsubscribe(ch2)
	assert.Len(t, b.subscribers, 0)
}

func TestBroadcastHandler(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b)

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// Test Handle
	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test log", 0)
	r.AddAttrs(slog.String("source", "test-source"))

	err := h.Handle(ctx, r)
	require.NoError(t, err)

	select {
	case data := <-ch:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test log", entry.Message)
		assert.Equal(t, "test-source", entry.Source)
		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Timestamp)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}

	// Test WithAttrs
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	assert.NotEqual(t, h, h2)
	// Just verify it doesn't panic and returns a handler
	assert.NotNil(t, h2)

	// Test WithGroup
	h3 := h.WithGroup("mygroup")
	assert.NotEqual(t, h, h3)
	assert.NotNil(t, h3)
}

func TestTeeHandler(t *testing.T) {
	// Mock handlers
	h1 := &mockHandler{}
	h2 := &mockHandler{}

	tee := NewTeeHandler(h1, h2)

	// Test Enabled
	ctx := context.Background()
	h1.enabled = true
	h2.enabled = false
	assert.True(t, tee.Enabled(ctx, slog.LevelInfo))

	h1.enabled = false
	assert.False(t, tee.Enabled(ctx, slog.LevelInfo))

	// Test Handle
	h1.enabled = true
	h2.enabled = true
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	err := tee.Handle(ctx, r)
	assert.NoError(t, err)
	assert.True(t, h1.handled)
	assert.True(t, h2.handled)

	// Test WithAttrs
	teeWithAttrs := tee.WithAttrs([]slog.Attr{slog.String("k", "v")})
	assert.NotNil(t, teeWithAttrs)
	assert.IsType(t, &TeeHandler{}, teeWithAttrs)
	// In a real mock we'd verify WithAttrs was called on children,
	// but for now we assume implementation is correct if it returns.

	// Test WithGroup
	teeWithGroup := tee.WithGroup("g")
	assert.NotNil(t, teeWithGroup)
	assert.IsType(t, &TeeHandler{}, teeWithGroup)
}

type mockHandler struct {
	enabled bool
	handled bool
}

func (m *mockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.enabled
}

func (m *mockHandler) Handle(ctx context.Context, r slog.Record) error {
	m.handled = true
	return nil
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	return m
}
