// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastHandler_WithAttrs(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// Add attributes via WithAttrs
	h2 := h.WithAttrs([]slog.Attr{slog.String("persistent_key", "persistent_value")})

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test log", 0)

	err := h2.Handle(ctx, r)
	require.NoError(t, err)

	select {
	case data := <-ch:
		entry, ok := data.(LogEntry)
		require.True(t, ok, "Expected data to be LogEntry")

		val, ok := entry.Metadata["persistent_key"]
		assert.True(t, ok, "persistent_key should exist in metadata")
		assert.Equal(t, "persistent_value", val)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}

func TestBroadcastHandler_WithGroup(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// Add group via WithGroup
	h2 := h.WithGroup("my_group")

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test log", 0)
	r.AddAttrs(slog.String("inner_key", "inner_value"))

	err := h2.Handle(ctx, r)
	require.NoError(t, err)

	select {
	case data := <-ch:
		entry, ok := data.(LogEntry)
		require.True(t, ok, "Expected data to be LogEntry")

		// If group is working, there should be a key "my_group" containing map/object
		// OR at least some structure.
		// If it is ignored, inner_key will be at root.

		_, ok = entry.Metadata["my_group"]
		// We assert that my_group should exist because that's the expected behavior
		assert.True(t, ok, "Metadata should contain group key 'my_group'")

		// And inner_key should be nested inside (if we implement nesting)
		// Or verify inner_key is NOT at root if we want strictness.
		_, innerOk := entry.Metadata["inner_key"]
		assert.False(t, innerOk, "inner_key should be nested, not at root")

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}
func TestBroadcastHandler_SourcePriority(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	// Case 1: Only component
	ctx := context.Background()
	r1 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg1", 0)
	r1.AddAttrs(slog.String("component", "my-component"))
	require.NoError(t, h.Handle(ctx, r1))

	select {
	case data := <-ch:
		entry, ok := data.(LogEntry)
		require.True(t, ok)
		assert.Equal(t, "my-component", entry.Source)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}

	// Case 2: source overrides component (per our deterministic choice)
	// Or at least source is picked if both present
	r2 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg2", 0)
	r2.AddAttrs(slog.String("component", "my-comp"), slog.String("source", "my-source"))
	require.NoError(t, h.Handle(ctx, r2))

	select {
	case data := <-ch:
		entry, ok := data.(LogEntry)
		require.True(t, ok)
		assert.Equal(t, "my-source", entry.Source)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}

	// Case 3: toolName overrides source
	r3 := slog.NewRecord(time.Now(), slog.LevelInfo, "msg3", 0)
	r3.AddAttrs(slog.String("source", "my-source"), slog.String("toolName", "my-tool"))
	require.NoError(t, h.Handle(ctx, r3))

	select {
	case data := <-ch:
		entry, ok := data.(LogEntry)
		require.True(t, ok)
		assert.Equal(t, "my-tool", entry.Source)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}
