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

func TestBroadcastHandler_WithAttrs_And_WithGroup(t *testing.T) {
	b := NewBroadcaster()
	// Cast to *BroadcastHandler to access specific methods if needed, but here we use interface methods
	var h slog.Handler = NewBroadcastHandler(b, slog.LevelInfo)

	// 1. Add top-level attribute
	h = h.WithAttrs([]slog.Attr{slog.String("app", "test-app")})

	// 2. Start a group "request"
	h = h.WithGroup("request")

	// 3. Add attribute inside group
	h = h.WithAttrs([]slog.Attr{slog.String("id", "req-123")})

	ch := b.Subscribe()
	defer b.Unsubscribe(ch)

	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "processing request", 0)

	// 4. Add attribute to the record (should be inside "request" group because handler has that group)
	r.AddAttrs(slog.String("status", "ok"))

	// Handle the record
	err := h.Handle(ctx, r)
	require.NoError(t, err)

	select {
	case data := <-ch:
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		t.Logf("Received Metadata: %+v", entry.Metadata)

		// Assertions for correct behavior (which currently fails)

		// "app" should be at top level
		assert.Equal(t, "test-app", entry.Metadata["app"], "Expected 'app' attribute at top level")

		// "request" should be a map (group)
		reqGroup, ok := entry.Metadata["request"].(map[string]interface{})
		if assert.True(t, ok, "Expected 'request' to be a group (map)") {
			// "id" should be inside "request"
			assert.Equal(t, "req-123", reqGroup["id"], "Expected 'id' inside 'request' group")
			// "status" should be inside "request"
			assert.Equal(t, "ok", reqGroup["status"], "Expected 'status' inside 'request' group")
		}

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}
