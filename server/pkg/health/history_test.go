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
	historyMu.Lock()
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu.Unlock()

	now := time.Now()
	// Helper to add history with timestamp
	add := func(svc string, ts int64, status string) {
		historyMu.Lock()
		defer historyMu.Unlock()
		hist, ok := historyStore[svc]
		if !ok {
			hist = &ServiceHealthHistory{Points: make([]HistoryPoint, 0, 100)}
			historyStore[svc] = hist
		}
		hist.Points = append(hist.Points, HistoryPoint{Timestamp: ts, Status: status})
	}

	svc := "test-svc"
	window := 1 * time.Hour

	// Case 1: No history
	assert.Equal(t, 0.0, CalculateUptime(svc, window))

	// Case 2: 100% Uptime (Started before window)
	// Point at T - 2h: UP
	add(svc, now.Add(-2*time.Hour).UnixMilli(), "UP")
	assert.Equal(t, 100.0, CalculateUptime(svc, window))

	// Case 3: 0% Uptime (Started before window)
	svc = "down-svc"
	add(svc, now.Add(-2*time.Hour).UnixMilli(), "DOWN")
	assert.Equal(t, 0.0, CalculateUptime(svc, window))

	// Case 4: Mixed (UP 30m, DOWN 30m)
	svc = "mixed-svc"
	// UP at T-1h (start of window)
	add(svc, now.Add(-1*time.Hour).UnixMilli(), "UP")
	// DOWN at T-30m
	add(svc, now.Add(-30*time.Minute).UnixMilli(), "DOWN")
	// Result: [T-1h, T-30m] UP (30m). [T-30m, T] DOWN (30m).
	// Uptime 50%
	assert.InDelta(t, 50.0, CalculateUptime(svc, window), 0.1)

	// Case 5: History starts within window
	svc = "short-history"
	// UP at T-30m.
	// Window starts at T-1h.
	// But history only starts at T-30m.
	// Interval [T-1h, T-30m] is ignored.
	// Interval [T-30m, T] is UP.
	// Total duration considered: 30m.
	// Uptime: 100%.
	add(svc, now.Add(-30*time.Minute).UnixMilli(), "UP")
	assert.Equal(t, 100.0, CalculateUptime(svc, window))
}
