// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// HistoryPoint represents the public HistoryPoint entity.
//
// Summary: Defines the structured data model representing a point.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type HistoryPoint struct {
	Timestamp int64  `json:"timestamp"` // Unix millis
	Status    string `json:"status"`
}

// ServiceHealthHistory represents the public ServiceHealthHistory entity.
//
// Summary: Defines the structured data model representing a health history.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type ServiceHealthHistory struct {
	Points []HistoryPoint
}

var (
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu    sync.RWMutex
)

// AddHealthStatus serves as a public interface for interacting with AddHealthStatus.
//
// Summary: Add the health status appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetHealthHistory serves as a public interface for interacting with GetHealthHistory.
//
// Summary: Fetches and returns the underlying health history from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
