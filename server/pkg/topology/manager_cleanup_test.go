// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"
)

func TestManager_cleanupSessions(t *testing.T) {
	// Create a manager (dependencies can be nil for this test as we manipulate sessions directly)
	m := &Manager{
		sessions: make(map[string]*SessionStats),
	}

	// 1. Add an old session (older than 24h)
	oldSessionID := "old-session"
	m.sessions[oldSessionID] = &SessionStats{
		ID:         oldSessionID,
		LastActive: time.Now().Add(-25 * time.Hour),
	}

	// 2. Add a new session (newer than 24h)
	newSessionID := "new-session"
	m.sessions[newSessionID] = &SessionStats{
		ID:         newSessionID,
		LastActive: time.Now().Add(-1 * time.Hour),
	}

	// 3. Run cleanup
	m.cleanupSessions()

	// 4. Verify old session is gone
	if _, exists := m.sessions[oldSessionID]; exists {
		t.Errorf("Old session %s should have been cleaned up", oldSessionID)
	}

	// 5. Verify new session remains
	if _, exists := m.sessions[newSessionID]; !exists {
		t.Errorf("New session %s should still exist", newSessionID)
	}
}
