// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateUptime(t *testing.T) {
	// Helper to reset store
	resetStore := func() {
		historyMu.Lock()
		historyStore = make(map[string]*ServiceHealthHistory)
		historyMu.Unlock()
	}

	t.Run("No History", func(t *testing.T) {
		resetStore()
		uptime := CalculateUptime("unknown_service", time.Hour)
		assert.Equal(t, 100.0, uptime)
	})

	t.Run("All UP in Window", func(t *testing.T) {
		resetStore()
		now := time.Now()
		// Inject history directly to control timestamps
		historyMu.Lock()
		historyStore["svc1"] = &ServiceHealthHistory{
			Points: []HistoryPoint{
				{Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "healthy"},
			},
		}
		historyMu.Unlock()

		uptime := CalculateUptime("svc1", time.Hour)
		assert.Equal(t, 100.0, uptime)
	})

	t.Run("All DOWN in Window", func(t *testing.T) {
		resetStore()
		now := time.Now()
		historyMu.Lock()
		historyStore["svc1"] = &ServiceHealthHistory{
			Points: []HistoryPoint{
				{Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "unhealthy"},
			},
		}
		historyMu.Unlock()

		uptime := CalculateUptime("svc1", time.Hour)
		assert.Equal(t, 0.0, uptime)
	})

	t.Run("Mixed History (50% UP)", func(t *testing.T) {
		resetStore()
		now := time.Now()
		// Window: 1 hour.
		// -1h: Starts UP (implied by first point being later? No, if first point is later, start is effective start)

		// Let's set points:
		// T-60m: UP
		// T-30m: DOWN
		// NOW: Still DOWN.
		// Expect 30m UP, 30m DOWN => 50%

		historyMu.Lock()
		historyStore["svc1"] = &ServiceHealthHistory{
			Points: []HistoryPoint{
				{Timestamp: now.Add(-60 * time.Minute).UnixMilli(), Status: "healthy"},
				{Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "unhealthy"},
			},
		}
		historyMu.Unlock()

		uptime := CalculateUptime("svc1", time.Hour)
		assert.InDelta(t, 50.0, uptime, 0.1)
	})

	t.Run("History Before Window", func(t *testing.T) {
		resetStore()
		now := time.Now()
		// Window: 1 hour.
		// T-2h: DOWN
		// T-1.5h: UP
		// Window start (T-1h): Status is UP.
		// T-30m: DOWN
		// NOW: Still DOWN.

		// In window (last 60m):
		// 0-30m: UP (from T-1.5h event)
		// 30-60m: DOWN (from T-30m event)
		// Expect 50%

		historyMu.Lock()
		historyStore["svc1"] = &ServiceHealthHistory{
			Points: []HistoryPoint{
				{Timestamp: now.Add(-120 * time.Minute).UnixMilli(), Status: "unhealthy"},
				{Timestamp: now.Add(-90 * time.Minute).UnixMilli(), Status: "healthy"},
				{Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "unhealthy"},
			},
		}
		historyMu.Unlock()

		uptime := CalculateUptime("svc1", time.Hour)
		assert.InDelta(t, 50.0, uptime, 0.1)
	})

	t.Run("Complex Mixed", func(t *testing.T) {
		resetStore()
		now := time.Now()
		// Window: 100m
		// T-90m: UP
		// T-80m: DOWN
		// T-20m: UP

		// Window [-100, 0]
		// Start (-100): Status unknown? No, effectively starts at T-90m if that's the first point.
		// Wait, if first point is inside window (T-90 > T-100), effective start is T-90.
		// Duration: 90m.
		// 0-10m (T-90 to T-80): UP. (10m)
		// 10-70m (T-80 to T-20): DOWN. (60m)
		// 70-90m (T-20 to T-0): UP. (20m)
		// Total UP: 30m. Total Time: 90m.
		// Uptime: 33.33%

		historyMu.Lock()
		historyStore["svc1"] = &ServiceHealthHistory{
			Points: []HistoryPoint{
				{Timestamp: now.Add(-90 * time.Minute).UnixMilli(), Status: "healthy"},
				{Timestamp: now.Add(-80 * time.Minute).UnixMilli(), Status: "unhealthy"},
				{Timestamp: now.Add(-20 * time.Minute).UnixMilli(), Status: "healthy"},
			},
		}
		historyMu.Unlock()

		uptime := CalculateUptime("svc1", 100*time.Minute)
		assert.InDelta(t, 33.33, uptime, 0.1)
	})
}
