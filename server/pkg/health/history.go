// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"sync"
	"time"
)

// HealthRecord represents a single health check result.
type HealthRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Status    bool      `json:"status"` // true = up, false = down
}

// HistoryManager manages the history of health checks for services.
type HistoryManager struct {
	mu      sync.RWMutex
	history map[string][]HealthRecord
	maxSize int
}

var (
	globalHistory *HistoryManager
	historyOnce   sync.Once
)

// GetHistoryManager returns the singleton HistoryManager.
func GetHistoryManager() *HistoryManager {
	historyOnce.Do(func() {
		globalHistory = &HistoryManager{
			history: make(map[string][]HealthRecord),
			maxSize: 1000, // Store last 1000 checks per service
		}
	})
	return globalHistory
}

// AddRecord adds a new health record for a service.
func (hm *HistoryManager) AddRecord(service string, status bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	record := HealthRecord{
		Timestamp: time.Now(),
		Status:    status,
	}

	hm.history[service] = append(hm.history[service], record)

	// Prune if exceeds max size
	if len(hm.history[service]) > hm.maxSize {
		overflow := len(hm.history[service]) - hm.maxSize
		hm.history[service] = hm.history[service][overflow:]
	}
}

// GetHistory returns the health history for a service.
func (hm *HistoryManager) GetHistory(service string) []HealthRecord {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if records, ok := hm.history[service]; ok {
		// Return a copy to avoid race conditions
		dst := make([]HealthRecord, len(records))
		copy(dst, records)
		return dst
	}
	return []HealthRecord{}
}
