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

func TestBroadcastHandler_WithAttrs_Behavior(t *testing.T) {
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

		// Verify persistent key is present
		val, ok := entry.Metadata["persistent_key"]
		assert.True(t, ok, "persistent_key should exist in metadata")
		assert.Equal(t, "persistent_value", val)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for broadcast")
	}
}
