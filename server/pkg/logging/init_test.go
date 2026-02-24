// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestGetLoggerBeforeInit(t *testing.T) {
	// Reset logger for test
	ForTestsOnlyResetLogger()

	// 1. Call GetLogger BEFORE Init (Triggers lazy default logger)
	logger1 := GetLogger()
	logger1.Info("Log before Init")

	// 2. Call Init (Should overwrite logger, but currently blocked by sync.Once)
	Init(slog.LevelInfo, os.Stderr, "", "text")

	// 3. GetLogger again
	logger2 := GetLogger()
	logger2.Info("Log after Init")

	// 4. Check if broadcaster received "Log after Init"
	// If Init was blocked, the logger is still the default stderr one (no broadcaster).
	// If Init worked, the logger has BroadcastHandler attached.

	ch, history := GlobalBroadcaster.SubscribeWithHistory()
	defer GlobalBroadcaster.Unsubscribe(ch)

	found := false
	for _, msg := range history {
		entry, ok := msg.(LogEntry)
		if ok && entry.Message == "Log after Init" {
			found = true
			break
		}
	}

	if !found {
		// Wait a bit
		select {
		case msg := <-ch:
			entry, ok := msg.(LogEntry)
			if ok && entry.Message == "Log after Init" {
				found = true
			}
		case <-time.After(100 * time.Millisecond):
		}
	}

	if !found {
		t.Log("Reproduction successful: Log message not found in broadcaster because Init was skipped")
        // Mark failed so I can see it fail before fix
        t.Fail()
	} else {
        t.Log("Log message FOUND. Bug NOT reproduced.")
    }
}
