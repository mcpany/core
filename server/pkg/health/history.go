// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// Record represents a single health check result.
type Record struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "UP" or "DOWN"
	LatencyMs int64     `json:"latency_ms"`
	Error     string    `json:"error,omitempty"`
}

// HistorySize is the number of records to keep per service.
const HistorySize = 1000

type serviceHistory struct {
	records []Record
	mu      sync.RWMutex
}

func newServiceHistory() *serviceHistory {
	return &serviceHistory{
		records: make([]Record, 0, HistorySize),
	}
}

func (h *serviceHistory) add(rec Record) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.records) >= HistorySize {
		// Shift
		h.records = h.records[1:]
	}
	h.records = append(h.records, rec)
}

func (h *serviceHistory) get() []Record {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]Record, len(h.records))
	copy(result, h.records)
	return result
}

var (
	globalHistory   = make(map[string]*serviceHistory)
	globalHistoryMu sync.RWMutex
)

// AddRecord adds a record to the history for a service.
func AddRecord(serviceName string, status string, latencyMs int64, errStr string) {
	globalHistoryMu.RLock()
	h, ok := globalHistory[serviceName]
	globalHistoryMu.RUnlock()

	if !ok {
		globalHistoryMu.Lock()
		// Double check
		if h, ok = globalHistory[serviceName]; !ok {
			h = newServiceHistory()
			globalHistory[serviceName] = h
		}
		globalHistoryMu.Unlock()
	}

	h.add(Record{
		Timestamp: time.Now(),
		Status:    status,
		LatencyMs: latencyMs,
		Error:     errStr,
	})
}

// GetHistory returns the health history for a service.
func GetHistory(serviceName string) []Record {
	globalHistoryMu.RLock()
	h, ok := globalHistory[serviceName]
	globalHistoryMu.RUnlock()

	if !ok {
		return []Record{}
	}
	return h.get()
}
