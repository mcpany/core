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

// CalculateUptime calculates the percentage of time a service was healthy within the given window.
func CalculateUptime(serviceName string, window time.Duration) float64 {
	historyMu.RLock()
	defer historyMu.RUnlock()

	hist, ok := historyStore[serviceName]
	if !ok || len(hist.Points) == 0 {
		return 0.0
	}

	now := time.Now()
	startTime := now.Add(-window)
	windowStart := startTime.UnixMilli()
	endTime := now.UnixMilli()

	var upDuration time.Duration
	var totalDuration time.Duration

	points := hist.Points

	for i := 0; i < len(points); i++ {
		var segStart, segEnd int64
		var status string

		segStart = points[i].Timestamp
		status = points[i].Status

		if i < len(points)-1 {
			segEnd = points[i+1].Timestamp
		} else {
			segEnd = endTime
		}

		// Intersect [segStart, segEnd] with [windowStart, endTime]
		overlapStart := maxInt64(segStart, windowStart)
		overlapEnd := minInt64(segEnd, endTime)

		if overlapEnd > overlapStart {
			duration := time.Duration(overlapEnd-overlapStart) * time.Millisecond
			totalDuration += duration
			if isHealthy(status) {
				upDuration += duration
			}
		}
	}

	if totalDuration == 0 {
		return 0
	}

	return (float64(upDuration) / float64(totalDuration)) * 100.0
}

func isHealthy(status string) bool {
	// Check against "up", "UP", "healthy"
	s := status
	return s == "up" || s == "UP" || s == "healthy" || s == "HEALTHY"
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
