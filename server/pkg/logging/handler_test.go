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

func TestBroadcastHandler_Enabled_Comprehensive(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)

	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
	assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
}

func TestBroadcastHandler_Handle_Comprehensive(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	logger := slog.New(h)
	msg := "test message"
	logger.Info(msg, "key", "value", "source", "test-source")

	select {
	case data := <-sub:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Timestamp)
		assert.Equal(t, "INFO", entry.Level)
		assert.Equal(t, msg, entry.Message)
		assert.Equal(t, "test-source", entry.Source)

		val, ok := entry.Metadata["key"]
		assert.True(t, ok)
		assert.Equal(t, "value", val)

		val, ok = entry.Metadata["source"]
		assert.True(t, ok)
		assert.Equal(t, "test-source", val)

	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log broadcast")
	}
}

func TestBroadcastHandler_SourcePriority(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	logger := slog.New(h)

	// Case 1: Default source (runtime caller)
	logger.Info("msg1")
	select {
	case data := <-sub:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(data, &entry))
		assert.NotEmpty(t, entry.Source)
		assert.Contains(t, entry.Source, "TestBroadcastHandler_SourcePriority")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log 1")
	}

	// Case 2: Explicit source overrides runtime caller
	logger.Info("msg2", "source", "explicit-source")
	select {
	case data := <-sub:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(data, &entry))
		assert.Equal(t, "explicit-source", entry.Source)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log 2")
	}

	// Case 3: Component overrides source
	logger.Info("msg3", "component", "my-component", "source", "my-source")
	select {
	case data := <-sub:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(data, &entry))
		assert.Equal(t, "my-component", entry.Source)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log 3")
	}

	// Case 4: toolName overrides component/source
	logger.Info("msg4", "component", "my-component", "toolName", "my-tool")
	select {
	case data := <-sub:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(data, &entry))
		assert.Equal(t, "my-tool", entry.Source)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log 4")
	}
}

func TestBroadcastHandler_WithAttrs_Verification(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	// Create a logger with pre-bound attributes
	logger := slog.New(h).With("env", "prod", "version", "1.0.0")
	logger.Info("test with attrs")

	select {
	case data := <-sub:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		assert.Equal(t, "test with attrs", entry.Message)

		val, ok := entry.Metadata["env"]
		assert.True(t, ok, "Metadata should contain 'env'")
		assert.Equal(t, "prod", val)

		val, ok = entry.Metadata["version"]
		assert.True(t, ok, "Metadata should contain 'version'")
		assert.Equal(t, "1.0.0", val)

	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log broadcast")
	}
}

func TestBroadcastHandler_WithGroup_Verification(t *testing.T) {
	b := NewBroadcaster()
	h := NewBroadcastHandler(b, slog.LevelInfo)
	sub := b.Subscribe()
	defer b.Unsubscribe(sub)

	logger := slog.New(h).WithGroup("http")
	logger.Info("request", "method", "GET", "path", "/api")

	select {
	case data := <-sub:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		val, ok := entry.Metadata["method"]
		assert.True(t, ok)
		assert.Equal(t, "GET", val)

	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for log broadcast")
	}
}

func TestTeeHandler_Comprehensive(t *testing.T) {
	b1 := NewBroadcaster()
	h1 := NewBroadcastHandler(b1, slog.LevelInfo)
	sub1 := b1.Subscribe()
	defer b1.Unsubscribe(sub1)

	b2 := NewBroadcaster()
	h2 := NewBroadcastHandler(b2, slog.LevelWarn) // Higher threshold
	sub2 := b2.Subscribe()
	defer b2.Unsubscribe(sub2)

	tee := NewTeeHandler(h1, h2)
	logger := slog.New(tee)

	// Test Info: Should go to h1 only
	logger.Info("info message")

	select {
	case <-sub1:
		// Received on h1
	case <-time.After(100 * time.Millisecond):
		t.Fatal("h1 should receive info")
	}

	select {
	case <-sub2:
		t.Fatal("h2 should NOT receive info")
	default:
		// OK
	}

	// Test Warn: Should go to both
	logger.Warn("warn message")

	select {
	case <-sub1:
		// Received on h1
	case <-time.After(100 * time.Millisecond):
		t.Fatal("h1 should receive warn")
	}

	select {
	case <-sub2:
		// Received on h2
	case <-time.After(100 * time.Millisecond):
		t.Fatal("h2 should receive warn")
	}
}

func TestTeeHandler_WithAttrs_Verification(t *testing.T) {
	b1 := NewBroadcaster()
	h1 := NewBroadcastHandler(b1, slog.LevelInfo)
	sub1 := b1.Subscribe()
	defer b1.Unsubscribe(sub1)

	tee := NewTeeHandler(h1)
	logger := slog.New(tee).With("tee_attr", "tee_val")

	logger.Info("msg")

	select {
	case data := <-sub1:
		var entry LogEntry
		require.NoError(t, json.Unmarshal(data, &entry))

		val, ok := entry.Metadata["tee_attr"]
		assert.True(t, ok)
		assert.Equal(t, "tee_val", val)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}
