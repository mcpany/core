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

func TestBroadcastHandler_WithAttrs_Repro(t *testing.T) {
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
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		// Verify that attributes from WithAttrs are present in metadata
		val, ok := entry.Metadata["persistent_key"]
		assert.True(t, ok, "persistent_key should exist in metadata")
		assert.Equal(t, "persistent_value", val)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}

func TestBroadcastHandler_WithGroup_Repro(t *testing.T) {
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
		var entry LogEntry
		err := json.Unmarshal(data, &entry)
		require.NoError(t, err)

		// Verify that the group is respected and creates a nested structure in metadata
		groupVal, ok := entry.Metadata["my_group"]
		assert.True(t, ok, "my_group should exist in metadata")

		groupMap, ok := groupVal.(map[string]interface{})
		assert.True(t, ok, "my_group should be a map")

		val, ok := groupMap["inner_key"]
		assert.True(t, ok, "inner_key should exist in my_group")
		assert.Equal(t, "inner_value", val)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}
