// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockToolManagerCleanup is a minimal mock for cleanup tests
type MockToolManagerCleanup struct {
	mock.Mock
	MockToolManager // Embed the full mock to avoid method mismatch errors if any interface changes
}

func TestManager_CleanupSessions(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)

	// Mock required calls during NewManager or subsequent operations if any
	// NewManager calls ListTools/ListServices? No, it just initializes maps.

	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// 1. Add recent session (active now)
	m.RecordActivity("recent-session", nil, 0, false, "")

	// Wait for recent session to be processed
	assert.Eventually(t, func() bool {
		m.mu.RLock()
		defer m.mu.RUnlock()
		_, exists := m.sessions["recent-session"]
		return exists
	}, 1*time.Second, 10*time.Millisecond, "Recent session should be processed")

	// 2. Add old session (active > sessionRetentionPeriod)
	// We need to manually inject this because RecordActivity sets LastActive to time.Now()
	m.mu.Lock()
	m.sessions["old-session"] = &SessionStats{
		ID:           "old-session",
		LastActive:   time.Now().Add(-(sessionRetentionPeriod + 1*time.Hour)),
		RequestCount: 1,
	}
	m.mu.Unlock()

	// Verify both exist
	m.mu.RLock()
	_, recentExists := m.sessions["recent-session"]
	_, oldExists := m.sessions["old-session"]
	m.mu.RUnlock()

	assert.True(t, recentExists, "Recent session should exist before cleanup")
	assert.True(t, oldExists, "Old session should exist before cleanup")

	// 3. Run Cleanup
	m.cleanupSessions()

	// 4. Verify old session is gone
	m.mu.RLock()
	_, recentExistsAfter := m.sessions["recent-session"]
	_, oldExistsAfter := m.sessions["old-session"]
	m.mu.RUnlock()

	assert.True(t, recentExistsAfter, "Recent session should remain after cleanup")
	assert.False(t, oldExistsAfter, "Old session should be removed after cleanup")
}
