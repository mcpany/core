// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
)

// initializeLogPersistence hydrates the logger with recent logs from storage.
func (a *Application) initializeLogPersistence(ctx context.Context, store storage.Storage) error {
	log := logging.GetLogger()
	logs, err := store.GetRecentLogs(ctx, 1000)
	if err != nil {
		log.Error("Failed to fetch recent logs for hydration", "error", err)
		return err
	}

	if len(logs) > 0 {
		history := make([]any, len(logs))
		for i, l := range logs {
			// logging.LogEntry is a struct, but Broadcaster expects any.
			// BroadcastHandler sends logging.LogEntry (struct).
			// store.GetRecentLogs returns pointers.
			history[i] = *l
		}
		logging.GlobalBroadcaster.Hydrate(history)
		log.Info("Hydrated log history from storage", "count", len(logs))
	}
	return nil
}

// startLogPersistence starts a background worker to persist logs to storage.
func (a *Application) startLogPersistence(ctx context.Context, store storage.Storage) {
	log := logging.GetLogger()
	// Create a buffered channel for log persistence to avoid blocking the broadcaster
	// We use a large buffer (e.g. 5000) to handle bursts.
	logCh := logging.GlobalBroadcaster.SubscribeBuffered(5000)

	go func() {
		defer logging.GlobalBroadcaster.Unsubscribe(logCh)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-logCh:
				if !ok {
					return
				}

				// Type assertion
				var entry logging.LogEntry
				if e, ok := msg.(logging.LogEntry); ok {
					entry = e
				} else if ptr, ok := msg.(*logging.LogEntry); ok {
					entry = *ptr
				} else {
					continue
				}

				// Save to DB
				// We use a separate context with timeout for DB operations to avoid hanging
				saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				// Note: If SaveLog fails, we avoid logging via standard logger to prevent infinite recursion loop
				_ = store.SaveLog(saveCtx, &entry)
				cancel()
			}
		}
	}()
	log.Info("Started log persistence worker")
}
