// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"strings"
	"sync"
)

// MemoryAuditStore stores audit logs in memory.
type MemoryAuditStore struct {
	mu      sync.RWMutex
	entries []Entry
}

// NewMemoryAuditStore creates a new MemoryAuditStore.
func NewMemoryAuditStore() *MemoryAuditStore {
	return &MemoryAuditStore{
		entries: make([]Entry, 0),
	}
}

// Write writes an audit entry to the store.
func (s *MemoryAuditStore) Write(_ context.Context, entry Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Prepend for newer first, or append? Typically logs are appended.
	// We'll append and filter/sort on read if needed, or reverse on read.
	s.entries = append(s.entries, entry)
	return nil
}

// Read reads audit entries from the store based on the filter.
func (s *MemoryAuditStore) Read(_ context.Context, filter Filter) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []Entry

	// Iterate in reverse order (newest first)
	for i := len(s.entries) - 1; i >= 0; i-- {
		entry := s.entries[i]

		if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
			continue
		}
		if filter.ToolName != "" && !strings.Contains(strings.ToLower(entry.ToolName), strings.ToLower(filter.ToolName)) {
			continue
		}
		if filter.UserID != "" && entry.UserID != filter.UserID {
			continue
		}
		if filter.ProfileID != "" && entry.ProfileID != filter.ProfileID {
			continue
		}

		result = append(result, entry)
	}

	// Pagination
	start := filter.Offset
	if start >= len(result) {
		return []Entry{}, nil
	}
	end := start + filter.Limit
	if filter.Limit == 0 {
		end = len(result)
	}
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

// Close closes the store.
func (s *MemoryAuditStore) Close() error {
	return nil
}
