package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanupSessions(t *testing.T) {
	// Create a manager with minimal dependencies (nil is fine for this test as we don't use them)
	m := &Manager{
		sessions: make(map[string]*SessionStats),
	}

	now := time.Now()

	// Add an active session
	m.sessions["active"] = &SessionStats{
		ID:         "active",
		LastActive: now,
	}

	// Add an expired session (older than 24h)
	m.sessions["expired"] = &SessionStats{
		ID:         "expired",
		LastActive: now.Add(-25 * time.Hour),
	}

	// Add a session exactly at cutoff (should stay or go? Before(cutoff) -> strict less than)
	// Let's test a session just inside the window
	m.sessions["boundary_active"] = &SessionStats{
		ID:         "boundary_active",
		LastActive: now.Add(-23 * time.Hour),
	}

	// Run cleanup with a cutoff of 24 hours ago
	cutoff := now.Add(-24 * time.Hour)
	m.cleanupSessions(cutoff)

	// Verify assertions
	assert.Contains(t, m.sessions, "active", "Active session should remain")
	assert.Contains(t, m.sessions, "boundary_active", "Boundary active session should remain")
	assert.NotContains(t, m.sessions, "expired", "Expired session should be removed")
	assert.Equal(t, 2, len(m.sessions), "Should have exactly 2 sessions remaining")
}
