// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_CleanupSessions(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	now := time.Now()

	// Add an old session (expired)
	m.mu.Lock()
	m.sessions["expired-session"] = &SessionStats{
		ID:         "expired-session",
		LastActive: now.Add(-25 * time.Hour),
	}
	// Add a new session (active)
	m.sessions["active-session"] = &SessionStats{
		ID:         "active-session",
		LastActive: now.Add(-1 * time.Hour),
	}
	m.mu.Unlock()

	// Call cleanup (this method will be added in the next step)
	m.cleanupSessions()

	m.mu.RLock()
	_, expiredExists := m.sessions["expired-session"]
	_, activeExists := m.sessions["active-session"]
	m.mu.RUnlock()

	assert.False(t, expiredExists, "Expired session should be removed")
	assert.True(t, activeExists, "Active session should remain")
}
