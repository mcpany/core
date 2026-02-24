// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// startLogPersistence starts a worker that persists logs to storage.
func (a *Application) startLogPersistence(ctx context.Context, store logging.LogPersister) {
	go func() {
		// Use a local logger to avoid infinite recursion if logging about logging fails?
		// Actually, since we subscribe to Broadcaster, if we log error here, it goes to Broadcaster -> channel -> here.
		// Infinite loop risk!
		// Solution: The logger inside this worker should NOT broadcast, or we should ignore errors from this worker in the subscription.
		// Or easier: Don't log errors from persistence using the global logger that broadcasts.
		// Use `fmt.Printf` or a separate logger?
		// Or relies on the fact that we process events. If `log.Error` generates an event, we receive it.
		// If `SaveLog` fails, we log Error. New event. We receive it. We try to save it. It fails. Loop.
		// Yes, infinite loop risk.

		// To avoid this, we can check the source or avoid logging errors to the main logger.
		// Or we can just print to stderr for persistence errors.

		// For now, I will use fmt.Println for errors in this worker.
		// Or I can check if the log message source is "log-persistence".

		ch := logging.GlobalBroadcaster.Subscribe()
		defer logging.GlobalBroadcaster.Unsubscribe(ch)

		// Batching
		batchSize := 50
		flushInterval := 2 * time.Second
		buffer := make([]logging.LogEntry, 0, batchSize)
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		flush := func() {
			if len(buffer) == 0 {
				return
			}
			for _, entry := range buffer {
				if err := store.SaveLog(ctx, entry); err != nil {
					// Write to stderr to avoid infinite logging loop
					fmt.Fprintf(os.Stderr, "Failed to persist log: %v\n", err)
				}
			}
			buffer = buffer[:0]
		}

		for {
			select {
			case <-ctx.Done():
				flush() // Flush remaining
				return
			case msg := <-ch:
				if entry, ok := msg.(logging.LogEntry); ok {
					buffer = append(buffer, entry)
					if len(buffer) >= batchSize {
						flush()
					}
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}
