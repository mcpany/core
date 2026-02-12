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
	// mockTM.On("ListTools").Return([]tool.Tool{}) // Not needed for this test
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// Add active session
	m.sessions["active"] = &SessionStats{
		ID:         "active",
		LastActive: time.Now(),
	}

	// Add expired session
	m.sessions["expired"] = &SessionStats{
		ID:         "expired",
		LastActive: time.Now().Add(-25 * time.Hour),
	}

	// Run cleanup with 24h cutoff
	cutoff := time.Now().Add(-24 * time.Hour)
	m.cleanupSessions(cutoff)

	m.mu.RLock()
	defer m.mu.RUnlock()

	_, activeExists := m.sessions["active"]
	_, expiredExists := m.sessions["expired"]

	assert.True(t, activeExists, "Active session should remain")
	assert.False(t, expiredExists, "Expired session should be removed")
}
