// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"
)

// TestManager_CleanupSessions tests the cleanupSessions method directly.
// We use the same package 'topology' to access unexported methods/fields.
func TestManager_CleanupSessions(t *testing.T) {
	// Create a minimal manager manually since NewManager starts a goroutine we don't need for this unit test.
	m := &Manager{
		sessions: make(map[string]*SessionStats),
	}

	// Create active session
	activeID := "active-session"
	m.sessions[activeID] = &SessionStats{
		ID:         activeID,
		LastActive: time.Now(),
	}

	// Create expired session (older than 24h)
	expiredID := "expired-session"
	m.sessions[expiredID] = &SessionStats{
		ID:         expiredID,
		LastActive: time.Now().Add(-25 * time.Hour),
	}

	// Create another expired session (exactly 25h)
	expiredID2 := "expired-session-2"
	m.sessions[expiredID2] = &SessionStats{
		ID:         expiredID2,
		LastActive: time.Now().Add(-25 * time.Hour),
	}

	// Run cleanup
	m.cleanupSessions()

	// Verify results
	if _, exists := m.sessions[activeID]; !exists {
		t.Errorf("Active session should remain")
	}

	if _, exists := m.sessions[expiredID]; exists {
		t.Errorf("Expired session should be removed")
	}

	if _, exists := m.sessions[expiredID2]; exists {
		t.Errorf("Expired session 2 should be removed")
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected 1 session remaining, got %d", len(m.sessions))
	}
}
