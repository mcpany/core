package topology

import (
	"testing"
	"time"
)

func TestManager_SessionCleanup(t *testing.T) {
	// ⚡ BOLT: Verify memory leak fix
	// Randomized Selection from Top 5 High-Impact Targets

	m := &Manager{
		sessions: make(map[string]*SessionStats),
	}

	now := time.Now()

	// Add an active session
	m.sessions["active"] = &SessionStats{
		ID:         "active",
		LastActive: now.Add(-1 * time.Hour),
	}

	// Add an expired session (older than 24h)
	m.sessions["expired"] = &SessionStats{
		ID:         "expired",
		LastActive: now.Add(-25 * time.Hour),
	}

	// Run cleanup
	m.cleanupSessions()

	// Verify active session remains
	if _, ok := m.sessions["active"]; !ok {
		t.Error("Active session was incorrectly removed")
	}

	// Verify expired session is removed
	if _, ok := m.sessions["expired"]; ok {
		t.Error("Expired session was not removed")
	}
}
