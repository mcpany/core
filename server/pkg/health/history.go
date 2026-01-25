// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// StatusEntry represents a health status change event.
type StatusEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	ServiceName string    `json:"service_name"`
}

// History manages the history of health status changes.
type History struct {
	mu      sync.Mutex
	entries []StatusEntry
	limit   int
}

var (
	globalHistory *History
	historyOnce   sync.Once
)

// GetHistory returns the global health history instance.
func GetHistory() *History {
	historyOnce.Do(func() {
		globalHistory = &History{
			limit: 100, // Keep last 100 entries
		}
	})
	return globalHistory
}

// Add records a new status change.
func (h *History) Add(serviceName, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	entry := StatusEntry{
		Timestamp:   time.Now(),
		Status:      status,
		ServiceName: serviceName,
	}

	h.entries = append(h.entries, entry)
	if len(h.entries) > h.limit {
		// Remove oldest
		h.entries = h.entries[1:]
	}
}

// Get returns a copy of the recorded history.
func (h *History) Get() []StatusEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Return copy
	result := make([]StatusEntry, len(h.entries))
	copy(result, h.entries)
	return result
}
