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

func TestBroadcastHandler_Enabled_Unit(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)

	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestBroadcastHandler_Handle(t *testing.T) {
	b := NewBroadcaster()
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	h := NewBroadcastHandler(b, slog.LevelInfo)

	now := time.Now()
	r := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	r.AddAttrs(slog.String("key", "value"))

	err := h.Handle(context.Background(), r)
	require.NoError(t, err)

	select {
	case msg := <-sub:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)

		assert.NotEmpty(t, entry.ID)
		// Check timestamp format roughly
		ts, err := time.Parse(time.RFC3339, entry.Timestamp)
		require.NoError(t, err)
		assert.WithinDuration(t, now, ts, time.Second)

		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test message", entry.Message)
		assert.Equal(t, "value", entry.Metadata["key"])
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast")
	}
}

func TestBroadcastHandler_SourcePriority(t *testing.T) {
	b := NewBroadcaster()
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	h := NewBroadcastHandler(b, slog.LevelInfo)

	tests := []struct {
		name           string
		attrs          []slog.Attr
		expectedSource string
	}{
		{
			name: "toolName priority",
			attrs: []slog.Attr{
				slog.String("toolName", "my-tool"),
				slog.String("component", "my-component"),
				slog.String("source", "my-source"),
			},
			expectedSource: "my-tool",
		},
		{
			name: "component priority",
			attrs: []slog.Attr{
				slog.String("component", "my-component"),
				slog.String("source", "my-source"),
			},
			expectedSource: "my-component", // component and source have priority 1, whichever comes first?
			// The code says:
			// if a.Key == "source" || a.Key == "component" {
			//     if sourcePriority < 1 { ... }
			// }
			// So first one wins.
		},
		{
			name: "source priority",
			attrs: []slog.Attr{
				slog.String("source", "my-source"),
			},
			expectedSource: "my-source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear channel
			for len(sub) > 0 {
				<-sub
			}

			r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
			r.AddAttrs(tt.attrs...)

			err := h.Handle(context.Background(), r)
			require.NoError(t, err)

			select {
			case msg := <-sub:
				var entry LogEntry
				err := json.Unmarshal(msg, &entry)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSource, entry.Source)
			case <-time.After(time.Second):
				t.Fatal("timeout")
			}
		})
	}
}

func TestBroadcastHandler_WithAttrs(t *testing.T) {
	b := NewBroadcaster()
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	h := NewBroadcastHandler(b, slog.LevelInfo)
	hWithAttrs := h.WithAttrs([]slog.Attr{slog.String("pre", "val")})

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	r.AddAttrs(slog.String("runtime", "val2"))

	err := hWithAttrs.Handle(context.Background(), r)
	require.NoError(t, err)

	select {
	case msg := <-sub:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)

		assert.Equal(t, "val", entry.Metadata["pre"])
		assert.Equal(t, "val2", entry.Metadata["runtime"])
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestTeeHandler_Unit(t *testing.T) {
	b1 := NewBroadcaster()
	sub1 := b1.Subscribe()
	defer b1.Unsubscribe(sub1)

	b2 := NewBroadcaster()
	sub2 := b2.Subscribe()
	defer b2.Unsubscribe(sub2)

	h1 := NewBroadcastHandler(b1, slog.LevelInfo)
	h2 := NewBroadcastHandler(b2, slog.LevelWarn)

	tee := NewTeeHandler(h1, h2)

	// Test Enabled
	assert.True(t, tee.Enabled(context.Background(), slog.LevelInfo)) // h1 is enabled
	assert.True(t, tee.Enabled(context.Background(), slog.LevelWarn)) // both enabled
	assert.False(t, tee.Enabled(context.Background(), slog.LevelDebug)) // neither enabled

	// Test Handle - Info (only h1)
	r1 := slog.NewRecord(time.Now(), slog.LevelInfo, "info msg", 0)
	err := tee.Handle(context.Background(), r1)
	require.NoError(t, err)

	select {
	case <-sub1:
		// received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for h1")
	}

	select {
	case <-sub2:
		t.Fatal("h2 should not receive info log")
	case <-time.After(100 * time.Millisecond):
		// ok
	}

	// Test Handle - Warn (both)
	r2 := slog.NewRecord(time.Now(), slog.LevelWarn, "warn msg", 0)
	err = tee.Handle(context.Background(), r2)
	require.NoError(t, err)

	select {
	case <-sub1:
		// received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for h1")
	}

	select {
	case <-sub2:
		// received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for h2")
	}
}

func TestTeeHandler_WithAttrs(t *testing.T) {
	// Mock handlers
	h1 := &mockHandler{}
	h2 := &mockHandler{}

	tee := NewTeeHandler(h1, h2)
	teeWithAttrs := tee.WithAttrs([]slog.Attr{slog.String("k", "v")})

	// Check if teeWithAttrs is a TeeHandler
	th, ok := teeWithAttrs.(*TeeHandler)
	assert.True(t, ok)

	// Since mockHandler.WithAttrs returns self, the handlers in the new TeeHandler should be the same objects
	assert.Len(t, th.handlers, 2)
	assert.Equal(t, h1, th.handlers[0])
	assert.Equal(t, h2, th.handlers[1])
}
