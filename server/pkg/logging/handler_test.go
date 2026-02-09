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

func TestBroadcastHandler_Enabled(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)

	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.True(t, h.Enabled(context.Background(), slog.LevelError))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestBroadcastHandler_Handle(t *testing.T) {
	b := NewBroadcaster()
	ch := b.Subscribe()

	h := NewBroadcastHandler(b, slog.LevelInfo)
	logger := slog.New(h)

	logger.Info("test message", "key", "value")

	select {
	case msg := <-ch:
		var entry LogEntry
		err := json.Unmarshal(msg, &entry)
		require.NoError(t, err)

		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Timestamp)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, "test message", entry.Message)
		assert.Equal(t, "value", entry.Metadata["key"])

		// Timestamp should be parseable
		ts, err := time.Parse(time.RFC3339, entry.Timestamp)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now(), ts, time.Second)

	case <-time.After(time.Second):
		t.Fatal("timeout waiting for log message")
	}
}

func TestBroadcastHandler_SourcePriority(t *testing.T) {
	b := NewBroadcaster()
	ch := b.Subscribe()

	h := NewBroadcastHandler(b, slog.LevelInfo)
	logger := slog.New(h)

	// Case 1: Runtime caller (implicit)
	// We can't easily test this without knowing the caller, but if no other source is provided,
	// it should fall back to something. However, the implementation only sets it if r.PC != 0.
	// slog.New(h) might set PC depending on options. Default logger usually captures PC.
	logger.Info("runtime caller test")
	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		// The function name will be something like "github.com/mcpany/core/server/pkg/logging.TestBroadcastHandler_SourcePriority"
		assert.Contains(t, entry.Source, "TestBroadcastHandler_SourcePriority")
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// Case 2: "source" attribute
	logger.Info("source attr test", "source", "my-source")
	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		assert.Equal(t, "my-source", entry.Source)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// Case 3: "component" attribute (same priority as source)
	logger.Info("component attr test", "component", "my-component")
	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		assert.Equal(t, "my-component", entry.Source)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// Case 4: "toolName" attribute (higher priority)
	logger.Info("toolName attr test", "source", "ignored", "toolName", "my-tool")
	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		assert.Equal(t, "my-tool", entry.Source)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestBroadcastHandler_WithAttrs(t *testing.T) {
	b := NewBroadcaster()
	ch := b.Subscribe()

	h := NewBroadcastHandler(b, slog.LevelInfo)
	h2 := h.WithAttrs([]slog.Attr{slog.String("common", "attr")})
	logger := slog.New(h2)

	logger.Info("with attrs test", "specific", "val")

	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		assert.Equal(t, "attr", entry.Metadata["common"])
		assert.Equal(t, "val", entry.Metadata["specific"])
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestBroadcastHandler_WithGroup(t *testing.T) {
	b := NewBroadcaster()
	ch := b.Subscribe()

	h := NewBroadcastHandler(b, slog.LevelInfo)
	h2 := h.WithGroup("myGroup")
	logger := slog.New(h2)

	logger.Info("with group test", "key", "val")

	select {
	case msg := <-ch:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(msg, &entry))
		// Note: BroadcastHandler currently does not implement grouping logic (nesting or prefixing),
		// so attributes under a group appear at the top level.
		// We verify that the attribute is present.
		assert.Equal(t, "val", entry.Metadata["key"])
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestTeeHandler(t *testing.T) {
	b1 := NewBroadcaster()
	ch1 := b1.Subscribe()
	h1 := NewBroadcastHandler(b1, slog.LevelInfo)

	b2 := NewBroadcaster()
	ch2 := b2.Subscribe()
	h2 := NewBroadcastHandler(b2, slog.LevelError)

	tee := NewTeeHandler(h1, h2)
	logger := slog.New(tee)

	// Test Info: should go to h1 only
	logger.Info("info msg")
	select {
	case <-ch1:
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout h1 info")
	}
	select {
	case <-ch2:
		t.Fatal("should not go to h2")
	default:
	}

	// Test Error: should go to both
	logger.Error("error msg")
	select {
	case <-ch1:
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout h1 error")
	}
	select {
	case <-ch2:
		// success
	case <-time.After(time.Second):
		t.Fatal("timeout h2 error")
	}
}
