// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_GetTrafficHistory_Extended(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	// Case 1: Empty history
	points := m.GetTrafficHistory()
	assert.Len(t, points, 60, "Should return 60 points even if empty")
	for _, p := range points {
		assert.Equal(t, int64(0), p.Total)
		assert.Equal(t, int64(0), p.Errors)
		assert.Equal(t, int64(0), p.Latency)
	}

	// Case 2: With data
	// Manually inject data into trafficHistory
	now := time.Now().Truncate(time.Minute)
	m.mu.Lock()
	m.trafficHistory[now.Unix()] = &MinuteStats{
		Requests: 10,
		Errors:   2,
		Latency:  500, // Total latency
	}
	m.mu.Unlock()

	points = m.GetTrafficHistory()
	assert.Len(t, points, 60)

	// The last point corresponds to the current minute (i=0)
	// Or close to it depending on clock skew during test execution.
	// Since we used Truncate(Minute), it should match exactly one point if "now" hasn't shifted minute.

	found := false
	for _, p := range points {
		if p.Time == now.Format("15:04") {
			assert.Equal(t, int64(10), p.Total)
			assert.Equal(t, int64(2), p.Errors)
			assert.Equal(t, int64(50), p.Latency, "Latency should be averaged")
			found = true
			break
		}
	}
	assert.True(t, found, "Should find the injected data point")
}

func TestManager_SeedTrafficHistory_Extended(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	seedPoints := []TrafficPoint{
		{
			Time:    "12:00",
			Total:   100,
			Errors:  5,
			Latency: 20, // Avg latency
		},
		{
			Time:    "invalid-time",
			Total:   10,
		},
	}

	m.SeedTrafficHistory(seedPoints)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the seeded point
	// We need to construct the expected key
	now := time.Now()
	parsedTime, _ := time.Parse("15:04", "12:00")
	expectedTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())

	stats, ok := m.trafficHistory[expectedTime.Unix()]
	require.True(t, ok, "Seeded point should exist in map")

	assert.Equal(t, int64(100), stats.Requests)
	assert.Equal(t, int64(5), stats.Errors)
	assert.Equal(t, int64(2000), stats.Latency, "Stored latency should be total (Avg * Requests)")

	// Invalid time should not be added
	// We can't easily check for absence without iterating or guessing keys,
	// but we can check count if we assume map was empty.
	// However, NewManager initializes empty map.
	assert.Len(t, m.trafficHistory, 1, "Only valid points should be added")
}

func TestManager_RecordActivity_Cleanup(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	// Add an old record (older than 24h)
	oldTime := time.Now().Add(-25 * time.Hour).Truncate(time.Minute).Unix()
	m.mu.Lock()
	m.trafficHistory[oldTime] = &MinuteStats{Requests: 1}

	// Add a session that is close to triggering cleanup (99 requests)
	m.sessions["trigger-session"] = &SessionStats{
		ID:           "trigger-session",
		RequestCount: 99,
		LastActive:   time.Now(),
	}
	m.mu.Unlock()

	// Trigger the 100th request
	m.RecordActivity("trigger-session", nil, 10*time.Millisecond, false)

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify old record is gone
	_, exists := m.trafficHistory[oldTime]
	assert.False(t, exists, "Old traffic history should be cleaned up")

	// Verify new record exists (current time)
	now := time.Now().Truncate(time.Minute).Unix()
	_, exists = m.trafficHistory[now]
	assert.True(t, exists, "Current activity should be recorded")
}

func TestManager_RecordActivity_NewSession(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	// Test that metadata with non-string values is handled gracefully
	meta := map[string]interface{}{
		"userAgent": "agent-1",
		"numeric": 123,
		"bool": true,
	}

	m.RecordActivity("new-session", meta, 50*time.Millisecond, false)

	m.mu.RLock()
	session, ok := m.sessions["new-session"]
	m.mu.RUnlock()

	require.True(t, ok)
	assert.Equal(t, "agent-1", session.Metadata["userAgent"])
	assert.NotContains(t, session.Metadata, "numeric")
	assert.NotContains(t, session.Metadata, "bool")
}

func TestManager_GetTrafficHistory_AvgCalculation(t *testing.T) {
    mockRegistry := new(MockServiceRegistry)
    mockTM := new(MockToolManager)
    m := NewManager(mockRegistry, mockTM)

    now := time.Now().Truncate(time.Minute)
    m.mu.Lock()
    // Case where requests > 0 but latency is small (integer division check)
    m.trafficHistory[now.Unix()] = &MinuteStats{
        Requests: 3,
        Latency:  4, // 4/3 = 1
    }
    m.mu.Unlock()

    points := m.GetTrafficHistory()

    for _, p := range points {
        if p.Time == now.Format("15:04") {
            assert.Equal(t, int64(1), p.Latency)
            return
        }
    }
    assert.Fail(t, "Point not found")
}
