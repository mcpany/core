// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"strings"
	"sync"
)

// Manager handles the lifecycle and retrieval of marketplace data.
type Manager struct {
	mu       sync.RWMutex
	registry []CommunityServer
}

// NewManager creates a new Marketplace Manager.
func NewManager() *Manager {
	return &Manager{
		registry: DefaultRegistry,
	}
}

// ListCommunityServers returns the list of community servers.
// It supports basic filtering.
func (m *Manager) ListCommunityServers(ctx context.Context) ([]CommunityServer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid mutation
	result := make([]CommunityServer, len(m.registry))
	copy(result, m.registry)

	return result, nil
}

// Sync updates the registry from external sources (e.g. GitHub Awesome List).
// For now, it's a no-op as we rely on the static registry.
func (m *Manager) Sync(ctx context.Context) error {
    // TODO: Implement fetching from https://raw.githubusercontent.com/punkpeye/awesome-mcp-servers/main/README.md
    // and merging with known schemas.
	return nil
}

// FindServerByName returns a server by name (case-insensitive).
func (m *Manager) FindServerByName(name string) (*CommunityServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.registry {
		if strings.EqualFold(s.Name, name) {
			return &s, true
		}
	}
	return nil, false
}
