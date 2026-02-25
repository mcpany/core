// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// HistoryPoint represents a single point in time for a service's health.
//
// Summary: Records the health status of a service at a specific timestamp.
type HistoryPoint struct {
	// Timestamp is the Unix timestamp in milliseconds.
	Timestamp int64 `json:"timestamp"`
	// Status is the health status string (e.g., "healthy", "unhealthy").
	Status string `json:"status"`
}

// ServiceHealthHistory stores the health history for a single service.
//
// Summary: Collection of historical health data points for a service.
type ServiceHealthHistory struct {
	// Points contains the list of health status points.
	Points []HistoryPoint
}

var (
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu    sync.RWMutex
)

// AddHealthStatus adds a new status point to the health history of a service.
//
// Summary: Appends a health status snapshot to the service's history.
//
// It maintains a maximum of 1000 history points, pruning the oldest if necessary.
//
// Parameters:
//   - serviceName (string): The name of the service.
//   - status (string): The current health status of the service.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Modifies the global history store.
//   - Allocates memory if the service is not yet tracked.
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

// GetHealthHistory retrieves the complete health history for all tracked services.
//
// Summary: Returns a snapshot of all service health histories.
//
// Returns:
//   - map[string][]HistoryPoint: A map where keys are service names and values are lists of history points.
//
// Side Effects:
//   - Allocates memory for the result map and slices (deep copy).
//   - Acquires a read lock on the global history store.
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
