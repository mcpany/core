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

	latencyStore = make(map[string]time.Duration)
	latencyMu    sync.RWMutex
)

// UpdateLatency updates the latest latency for a service.
func UpdateLatency(serviceName string, d time.Duration) {
	latencyMu.Lock()
	defer latencyMu.Unlock()
	latencyStore[serviceName] = d
}

// GetLatestLatency returns the latest latency for a service.
func GetLatestLatency(serviceName string) time.Duration {
	latencyMu.RLock()
	defer latencyMu.RUnlock()
	return latencyStore[serviceName]
}

// CalculateUptime calculates the uptime percentage for a service over a given window.
func CalculateUptime(serviceName string, window time.Duration) float64 {
	historyMu.RLock()
	defer historyMu.RUnlock()

	hist, ok := historyStore[serviceName]
	if !ok || len(hist.Points) == 0 {
		return 0.0
	}

	now := time.Now()
	startTime := now.Add(-window)
	startMillis := startTime.UnixMilli()
	nowMillis := now.UnixMilli()

	// Find the state at the start of the window
	idx := -1
	for i, p := range hist.Points {
		if p.Timestamp >= startMillis {
			idx = i
			break
		}
	}

	var currentState string
	const statusHealthy = "healthy"

	isHealthy := func(status string) bool {
		return status == "up" || status == "UP" || status == statusHealthy
	}

	if idx == -1 {
		// All points are before startMillis.
		// State is the last point's status.
		lastPoint := hist.Points[len(hist.Points)-1]
		currentState = lastPoint.Status
		if isHealthy(currentState) {
			return 100.0
		}
		return 0.0
	}

	// Determine initial state at startMillis
	if idx > 0 {
		currentState = hist.Points[idx-1].Status
	} else {
		// No point before startMillis. Assume state of first point.
		currentState = hist.Points[0].Status
	}

	totalUpTime := int64(0)
	lastTime := startMillis

	// Iterate through points within the window
	for i := idx; i < len(hist.Points); i++ {
		p := hist.Points[i]
		// Add duration of previous state
		duration := p.Timestamp - lastTime
		if isHealthy(currentState) {
			totalUpTime += duration
		}
		lastTime = p.Timestamp
		currentState = p.Status
	}

	// Add duration from last point to now
	duration := nowMillis - lastTime
	if isHealthy(currentState) {
		totalUpTime += duration
	}

	totalDuration := nowMillis - startMillis
	if totalDuration == 0 {
		return 0.0
	}

	return float64(totalUpTime) / float64(totalDuration) * 100.0
}

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
