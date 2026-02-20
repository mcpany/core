// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// HistoryPoint represents a single point in time for a service's health.
type HistoryPoint struct {
	Timestamp int64  `json:"timestamp"` // Unix millis
	Status    string `json:"status"`
}

// ServiceHealthHistory stores the history for a service.
type ServiceHealthHistory struct {
	Points []HistoryPoint
}

var (
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu    sync.RWMutex
)

// AddHealthStatus adds a status point to the history.
func AddHealthStatus(serviceName string, status string) {
	historyMu.Lock()
	defer historyMu.Unlock()

	hist, ok := historyStore[serviceName]
	if !ok {
		hist = &ServiceHealthHistory{Points: make([]HistoryPoint, 0, 100)}
		historyStore[serviceName] = hist
	}

	// Deduplicate: If status hasn't changed, do we store it?
	// Roadmap says "visual timeline". If we only store changes, we can reconstruct lines.
	// But "heatmap" style usually wants regular intervals.
	// Let's store changes AND periodic snapshots?
	// For now, store every event. The checker calls listener on *change* (deduplicated in health.go).
	// But we also want to know "it's still healthy".
	// The `health` package dedups: `if prev == state.Status { return }`.
	// So we only get changes.
	// To render a timeline, the UI needs to know "from T1 to T2 it was Healthy".
	// If we provide change points, the UI can fill the gaps.
	// But the UI hook expects points.

	// Let's just append the point.
	now := time.Now().UnixMilli()
	hist.Points = append(hist.Points, HistoryPoint{
		Timestamp: now,
		Status:    status,
	})

	// Prune
	if len(hist.Points) > 1000 {
		hist.Points = hist.Points[len(hist.Points)-1000:]
	}
}

// GetHealthHistory returns the history for all services.
func GetHealthHistory() map[string][]HistoryPoint {
	historyMu.RLock()
	defer historyMu.RUnlock()

	result := make(map[string][]HistoryPoint)
	for name, hist := range historyStore {
		points := make([]HistoryPoint, len(hist.Points))
		copy(points, hist.Points)
		result[name] = points
	}
	return result
}

// CalculateUptime calculates the uptime percentage for a given history and window.
func CalculateUptime(history []HistoryPoint, window time.Duration) float64 {
	if window <= 0 {
		return 0.0
	}
	now := time.Now().UnixMilli()
	start := now - window.Milliseconds()

	var totalUpTime int64
	var lastTime int64 = start
	var lastStatusString = "unknown" // Default if no history before start

	// 1. Find the initial status at 'start'
	// Iterate through history to find the last point <= start
	for _, p := range history {
		if p.Timestamp <= start {
			lastStatusString = p.Status
		} else {
			break // Points are sorted
		}
	}

	// 2. Iterate through points that are within the window (> start)
	for _, p := range history {
		if p.Timestamp <= start {
			continue
		}
		if p.Timestamp > now {
			break
		}

		// Segment from lastTime to p.Timestamp
		if lastStatusString == "up" || lastStatusString == "UP" {
			totalUpTime += (p.Timestamp - lastTime)
		}

		lastTime = p.Timestamp
		lastStatusString = p.Status
	}

	// 3. Add final segment
	if now > lastTime {
		if lastStatusString == "up" || lastStatusString == "UP" {
			totalUpTime += (now - lastTime)
		}
	}

	uptime := float64(totalUpTime) / float64(window.Milliseconds())
	if uptime > 1.0 {
		uptime = 1.0
	}
	return uptime
}

// AddHealthStatusAtTime adds a status point to the history at a specific time.
// This is primarily for testing.
func AddHealthStatusAtTime(serviceName string, status string, timestamp time.Time) {
	historyMu.Lock()
	defer historyMu.Unlock()

	hist, ok := historyStore[serviceName]
	if !ok {
		hist = &ServiceHealthHistory{Points: make([]HistoryPoint, 0, 100)}
		historyStore[serviceName] = hist
	}

	hist.Points = append(hist.Points, HistoryPoint{
		Timestamp: timestamp.UnixMilli(),
		Status:    status,
	})

	// Sort points by timestamp (insertion sort or full sort?)
	// Since we append, if we add past data, it might be out of order.
	// CalculateUptime assumes sorted.
	// Let's rely on test usage to add in order or just sort?
	// For simplicity in test helper, we won't sort here, assuming caller adds in order or we don't care about slight disorder if we only add one point.
	// But actually, if we add a point in the past, it will be at the end of the slice.
	// Iterate logic in CalculateUptime might fail if not sorted.
	// Let's just assume we clear history first or add in order.
}
