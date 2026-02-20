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
//
// points: The history points.
// window: The time window to look back.
//
// Returns:
//   - uptime: The uptime percentage (0.0 - 1.0).
func CalculateUptime(points []HistoryPoint, window time.Duration) float64 {
	if len(points) == 0 {
		return 1.0
	}

	now := time.Now().UnixMilli()
	start := now - window.Milliseconds()
	if start < 0 {
		start = 0
	}

	var upDuration int64
	var lastTime = start
	var currentState = "unknown"

	// Find initial state at `start`
	// Iterate backwards to find the last point before start
	for i := len(points) - 1; i >= 0; i-- {
		if points[i].Timestamp <= start {
			currentState = points[i].Status
			break
		}
	}

	// If no point before start, use the first point's status
	if currentState == "unknown" && len(points) > 0 {
		currentState = points[0].Status
	}

	for _, p := range points {
		if p.Timestamp < start {
			continue
		}
		if p.Timestamp > now {
			break
		}

		// Calculate duration from lastTime to p.Timestamp
		duration := p.Timestamp - lastTime
		if duration > 0 {
			if isUp(currentState) {
				upDuration += duration
			}
		}

		lastTime = p.Timestamp
		currentState = p.Status
	}

	// Add duration from last point to now
	if now > lastTime {
		duration := now - lastTime
		if isUp(currentState) {
			upDuration += duration
		}
	}

	if window.Milliseconds() == 0 {
		return 0.0
	}

	return float64(upDuration) / float64(window.Milliseconds())
}

func isUp(status string) bool {
	return status == "up" || status == "UP" || status == "healthy" || status == "HEALTHY"
}
