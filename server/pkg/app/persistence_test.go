// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogPersistence(t *testing.T) {
	// Reset Global Broadcaster
	logging.GlobalBroadcaster = logging.NewBroadcaster()
	// Reset Logger
	logging.ForTestsOnlyResetLogger()

	store := memory.NewStore()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Simulate the worker logic that we added to server.go
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	// Subscribe synchronously to ensure we don't miss the broadcast
	logCh := logging.GlobalBroadcaster.Subscribe()

	go func() {
		defer logging.GlobalBroadcaster.Unsubscribe(logCh)

		for {
			select {
			case <-workerCtx.Done():
				return
			case msg := <-logCh:
				var entry storage.LogEntry
				if err := json.Unmarshal(msg, &entry); err == nil {
					_ = store.SaveLog(context.Background(), &entry)
				}
			}
		}
	}()

	// 2. Broadcast a log (Simulate logging.GetLogger().Info(...))
	// We manually broadcast bytes to simulate what BroadcastHandler does
	testLogJSON := `{"id": "1", "message": "test log", "level": "INFO", "timestamp": "2023-01-01T00:00:00Z"}`
	logging.GlobalBroadcaster.Broadcast([]byte(testLogJSON))

	// 3. Wait and verify persistence
	require.Eventually(t, func() bool {
		logs, _ := store.ListLogs(context.Background(), 100)
		return len(logs) > 0
	}, 2*time.Second, 10*time.Millisecond, "Log should be persisted to store")

	logs, err := store.ListLogs(context.Background(), 100)
	require.NoError(t, err)
	require.NotEmpty(t, logs)
	assert.Equal(t, "test log", logs[0].Message)

	// 4. Test Hydration (Simulate Server Restart)
	// Create a new broadcaster (simulating new server process)
	newBroadcaster := logging.NewBroadcaster()
	// Temporarily swap global to test HydrateFromStorage which uses GlobalBroadcaster
	old := logging.GlobalBroadcaster
	logging.GlobalBroadcaster = newBroadcaster
	defer func() { logging.GlobalBroadcaster = old }()

	// Hydrate from the store which has the log
	err = logging.HydrateFromStorage(context.Background(), store)
	require.NoError(t, err)

	// Verify history in the new broadcaster
	history := newBroadcaster.GetHistory()
	assert.NotEmpty(t, history, "Broadcaster history should be hydrated")

	var entry storage.LogEntry
	err = json.Unmarshal(history[0], &entry)
	require.NoError(t, err)
	assert.Equal(t, "test log", entry.Message)
}
