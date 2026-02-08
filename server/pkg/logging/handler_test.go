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

func TestBroadcastHandler_Enabled_Coverage(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)

	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h.Enabled(context.Background(), slog.LevelError))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestBroadcastHandler_Handle(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelDebug)
	ch := b.Subscribe()

	ctx := context.Background()
	timestamp := time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC)
	record := slog.NewRecord(timestamp, slog.LevelInfo, "test message", 0)
	record.AddAttrs(
		slog.String("key1", "value1"),
		slog.Int("key2", 123),
		slog.String("source", "my-source"),
	)

	err := h.Handle(ctx, record)
	require.NoError(t, err)

	select {
	case msg := <-ch:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)

		assert.NotEmpty(t, entry.ID)
		assert.Equal(t, timestamp.Format(time.RFC3339), entry.Timestamp)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test message", entry.Message)
		assert.Equal(t, "my-source", entry.Source) // Priority 1

		require.NotNil(t, entry.Metadata)
		assert.Equal(t, "value1", entry.Metadata["key1"])
		// JSON unmarshals numbers as float64 by default
		assert.Equal(t, float64(123), entry.Metadata["key2"])

	case <-time.After(time.Second):
		t.Fatal("timeout waiting for log message")
	}
}

func TestBroadcastHandler_SourcePriority(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelDebug)
	ch := b.Subscribe()

	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	record.AddAttrs(
		slog.String("source", "low-priority"),
		slog.String("toolName", "high-priority"),
	)

	err := h.Handle(ctx, record)
	require.NoError(t, err)

	select {
	case msg := <-ch:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)
		assert.Equal(t, "high-priority", entry.Source)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestBroadcastHandler_WithAttrs(t *testing.T) {
	b := NewBroadcaster()
	baseHandler := NewBroadcastHandler(b, slog.LevelInfo)
	h := baseHandler.WithAttrs([]slog.Attr{
		slog.String("base_attr", "base_val"),
	})
	ch := b.Subscribe()

	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	record.AddAttrs(slog.String("record_attr", "record_val"))

	err := h.Handle(ctx, record)
	require.NoError(t, err)

	select {
	case msg := <-ch:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)

		require.NotNil(t, entry.Metadata)
		assert.Equal(t, "record_val", entry.Metadata["record_attr"])

		// If implementation is correct, base_attr should be present
		val, ok := entry.Metadata["base_attr"]
		require.True(t, ok, "base_attr missing from Metadata")
		assert.Equal(t, "base_val", val)

	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

// MockHandler for TeeHandler tests
type MockHandler struct {
	EnabledFunc   func(context.Context, slog.Level) bool
	HandleFunc    func(context.Context, slog.Record) error
	WithAttrsFunc func([]slog.Attr) slog.Handler
	WithGroupFunc func(string) slog.Handler
}

func (m *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if m.EnabledFunc != nil {
		return m.EnabledFunc(ctx, level)
	}
	return true
}

func (m *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(ctx, r)
	}
	return nil
}

func (m *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if m.WithAttrsFunc != nil {
		return m.WithAttrsFunc(attrs)
	}
	return m
}

func (m *MockHandler) WithGroup(name string) slog.Handler {
	if m.WithGroupFunc != nil {
		return m.WithGroupFunc(name)
	}
	return m
}

func TestTeeHandler_Coverage(t *testing.T) {
	h1 := &MockHandler{
		EnabledFunc: func(_ context.Context, _ slog.Level) bool { return true },
		HandleFunc: func(_ context.Context, _ slog.Record) error { return nil },
	}
	h2 := &MockHandler{
		EnabledFunc: func(_ context.Context, _ slog.Level) bool { return false },
		HandleFunc: func(_ context.Context, _ slog.Record) error { return nil },
	}

	tee := NewTeeHandler(h1, h2)

	// Test Enabled
	assert.True(t, tee.Enabled(context.Background(), slog.LevelInfo))

	// Test Handle
	called1 := false
	h1.HandleFunc = func(_ context.Context, _ slog.Record) error {
		called1 = true
		return nil
	}
	called2 := false
	h2.HandleFunc = func(_ context.Context, _ slog.Record) error {
		called2 = true
		return nil
	}
	// h2 is disabled, so it shouldn't be called? Wait, tee.Enabled calls handlers.Enabled.
	// But Handle calls handlers.Enabled check too?
	// TeeHandler.Handle implementation:
	/*
		func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
			var err error
			for _, handler := range h.handlers {
				if handler.Enabled(ctx, r.Level) { // <--- Checks Enabled here
					if e := handler.Handle(ctx, r); e != nil {
						err = e
					}
				}
			}
			return err
		}
	*/

	err := tee.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0))
	require.NoError(t, err)
	assert.True(t, called1)
	assert.False(t, called2) // Because h2.Enabled returns false
}

func TestTeeHandler_WithAttrs(t *testing.T) {
	called1 := false
	h1 := &MockHandler{
		WithAttrsFunc: func(attrs []slog.Attr) slog.Handler {
			called1 = true
			return &MockHandler{}
		},
	}

	tee := NewTeeHandler(h1)
	_ = tee.WithAttrs([]slog.Attr{slog.String("a", "b")})
	assert.True(t, called1)
}
