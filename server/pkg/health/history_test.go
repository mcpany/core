// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthHistory(t *testing.T) {
	// Reset store
	historyMu.Lock()
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu.Unlock()

	AddHealthStatus("svc1", "healthy")
	AddHealthStatus("svc1", "unhealthy")

	hist := GetHealthHistory()
	assert.Len(t, hist["svc1"], 2)
	assert.Equal(t, "healthy", hist["svc1"][0].Status)
	assert.Equal(t, "unhealthy", hist["svc1"][1].Status)

	// Test pruning
	for i := 0; i < 1100; i++ {
		AddHealthStatus("svc2", "ok")
	}
	hist = GetHealthHistory()
	assert.Len(t, hist["svc2"], 1000)
}

func TestCalculateUptime(t *testing.T) {
	now := time.Now().UnixMilli()

	// Case 1: All uptime (started before window)
	history1 := []HistoryPoint{
		{Timestamp: now - 3600*1000, Status: "UP"}, // 1 hour ago
	}
	uptime1 := CalculateUptime(history1, 30*time.Minute)
	// Window is 30m. Status was UP 1h ago. So it was UP for the whole 30m.
	assert.InDelta(t, 1.0, uptime1, 0.001)

	// Case 2: Downtime in the middle
	// Window: 10 mins (600s)
	// Start: T-10m
	// T-10m: UP (implied by T-15m point)
	// T-5m: DOWN
	// T-2m: UP
	// End: T
	// Uptime: [T-10m, T-5m] = 5m UP. [T-5m, T-2m] = 3m DOWN. [T-2m, T] = 2m UP.
	// Total UP = 5 + 2 = 7m.
	// Percentage = 7/10 = 0.7.

	history2 := []HistoryPoint{
		{Timestamp: now - 15*60*1000, Status: "UP"},
		{Timestamp: now - 5*60*1000, Status: "DOWN"},
		{Timestamp: now - 2*60*1000, Status: "UP"},
	}
	uptime2 := CalculateUptime(history2, 10*time.Minute)
	assert.InDelta(t, 0.7, uptime2, 0.001)

	// Case 3: No history
	uptime3 := CalculateUptime([]HistoryPoint{}, 10*time.Minute)
	assert.Equal(t, 0.0, uptime3)

	// Case 4: History starts AFTER window start
	// Window: 10 mins
	// T-5m: UP
	// Interval [T-10m, T-5m] is unknown (not counted as UP).
	// Interval [T-5m, T] is UP.
	// Total UP = 5m.
	// Percentage = 0.5.
	history4 := []HistoryPoint{
		{Timestamp: now - 5*60*1000, Status: "UP"},
	}
	uptime4 := CalculateUptime(history4, 10*time.Minute)
	assert.InDelta(t, 0.5, uptime4, 0.001)
}
