// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"sync"
	"time"
)

// ExecutionRecord represents a single tool execution.
type ExecutionRecord struct {
	ToolName  string    `json:"tool_name"`
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// ExecutionHistory manages a ring buffer of execution records.
type ExecutionHistory struct {
	mu      sync.RWMutex
	records []ExecutionRecord
	size    int
	head    int // points to the next write position
	full    bool
}

// NewExecutionHistory creates a new history manager.
func NewExecutionHistory(size int) *ExecutionHistory {
	if size <= 0 {
		size = 100 // Default size
	}
	return &ExecutionHistory{
		records: make([]ExecutionRecord, size),
		size:    size,
	}
}

// Add adds a record to the history.
func (h *ExecutionHistory) Add(record ExecutionRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.records[h.head] = record
	h.head++
	if h.head >= h.size {
		h.head = 0
		h.full = true
	}
}

// List returns the records in chronological order (oldest to newest).
func (h *ExecutionHistory) List() []ExecutionRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []ExecutionRecord
	if !h.full {
		result = make([]ExecutionRecord, h.head)
		copy(result, h.records[:h.head])
	} else {
		result = make([]ExecutionRecord, h.size)
		// Oldest is at h.head
		copy(result, h.records[h.head:])
		copy(result[h.size-h.head:], h.records[:h.head])
	}
	return result
}
