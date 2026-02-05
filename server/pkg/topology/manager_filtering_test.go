// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_Filtering(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// Record activity for Service A
	m.RecordActivity("session-1", nil, 100*time.Millisecond, false, "service-a")

	// Record activity for Service B
	m.RecordActivity("session-2", nil, 200*time.Millisecond, true, "service-b")

	// Record activity with no service
	m.RecordActivity("session-3", nil, 50*time.Millisecond, false, "")

	// Wait for all 3 requests to be processed
	assert.Eventually(t, func() bool {
		globalStats := m.GetStats("")
		return globalStats.TotalRequests == 3
	}, 1*time.Second, 10*time.Millisecond)

	// Test Global Stats
	globalStats := m.GetStats("")
	assert.Equal(t, int64(3), globalStats.TotalRequests)
	// Latency: (100 + 200 + 50) / 3 = 116.66
	assert.Equal(t, 350*time.Millisecond/3, globalStats.AvgLatency)
	// Errors: 1 / 3
	assert.InDelta(t, 0.333, globalStats.ErrorRate, 0.01)

	// Test Service A Stats
	statsA := m.GetStats("service-a")
	assert.Equal(t, int64(1), statsA.TotalRequests)
	assert.Equal(t, 100*time.Millisecond, statsA.AvgLatency)
	assert.Equal(t, 0.0, statsA.ErrorRate)

	// Test Service B Stats
	statsB := m.GetStats("service-b")
	assert.Equal(t, int64(1), statsB.TotalRequests)
	assert.Equal(t, 200*time.Millisecond, statsB.AvgLatency)
	assert.Equal(t, 1.0, statsB.ErrorRate)

	// Test Traffic History Filtering
	historyA := m.GetTrafficHistory("service-a")
	if assert.NotEmpty(t, historyA) {
		lastPoint := historyA[len(historyA)-1]
		assert.Equal(t, int64(1), lastPoint.Total)
		assert.Equal(t, int64(0), lastPoint.Errors)
		assert.Equal(t, int64(100), lastPoint.Latency)
	}

	historyB := m.GetTrafficHistory("service-b")
	if assert.NotEmpty(t, historyB) {
		lastPoint := historyB[len(historyB)-1]
		assert.Equal(t, int64(1), lastPoint.Total)
		assert.Equal(t, int64(1), lastPoint.Errors)
		assert.Equal(t, int64(200), lastPoint.Latency)
	}
}
