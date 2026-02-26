// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"strings"
	"sync"

	v1 "github.com/mcpany/core/proto/api/v1"
	"google.golang.org/protobuf/proto"
)

// Manager manages the lazy tool index.
//
// Summary: Manages the index of tools available for lazy loading.
type Manager struct {
	mu          sync.RWMutex
	tools       []*v1.IndexedTool
	searchCount int32
	hits        int32
	misses      int32
}

// NewManager creates a new Index Manager.
//
// Summary: Initializes a new Index Manager.
//
// Returns:
//   - *Manager: The initialized manager.
func NewManager() *Manager {
	return &Manager{
		tools: make([]*v1.IndexedTool, 0),
	}
}

// Seed populates the index with tools.
//
// Summary: Seeds the index with test data.
//
// Parameters:
//   - _ (context.Context): The context (unused).
//   - tools ([]*v1.IndexedTool): The tools to seed.
//   - shouldClear (bool): Whether to clear existing data.
//
// Returns:
//   - int32: The number of tools seeded.
func (m *Manager) Seed(_ context.Context, tools []*v1.IndexedTool, shouldClear bool) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if shouldClear {
		m.tools = make([]*v1.IndexedTool, 0)
	}

	m.tools = append(m.tools, tools...)
	return int32(len(tools)) //nolint:gosec // Length unlikely to exceed int32
}

// Search searches the index for tools matching the query.
//
// Summary: Searches for tools.
//
// Parameters:
//   - _ (context.Context): The context (unused).
//   - query (string): The search query.
//   - page (int32): Page number (1-based).
//   - limit (int32): Page size.
//
// Returns:
//   - []*v1.IndexedTool: Matching tools.
//   - int32: Total matching count.
func (m *Manager) Search(_ context.Context, query string, page, limit int32) ([]*v1.IndexedTool, int32) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.searchCount++

	var results []*v1.IndexedTool
	q := strings.ToLower(query)

	for _, t := range m.tools {
		if q == "" ||
			strings.Contains(strings.ToLower(t.GetName()), q) ||
			strings.Contains(strings.ToLower(t.GetDescription()), q) ||
			containsTag(t.GetTags(), q) {
			results = append(results, t)
		}
	}

	total := int32(len(results)) //nolint:gosec // Length unlikely to exceed int32
	if total > 0 {
		m.hits++
	} else {
		m.misses++
	}

	// Pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []*v1.IndexedTool{}, total
	}
	if end > total {
		end = total
	}

	return results[start:end], total
}

func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// GetStats returns usage statistics.
//
// Summary: Gets index statistics.
//
// Parameters:
//   - _ (context.Context): The context (unused).
//
// Returns:
//   - *v1.GetIndexStatsResponse: The stats.
func (m *Manager) GetStats(_ context.Context) *v1.GetIndexStatsResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &v1.GetIndexStatsResponse{
		TotalTools:    proto.Int32(int32(len(m.tools))), //nolint:gosec // Length unlikely to exceed int32
		TotalSearches: proto.Int32(m.searchCount),
		Hits:          proto.Int32(m.hits),
		Misses:        proto.Int32(m.misses),
	}
}
