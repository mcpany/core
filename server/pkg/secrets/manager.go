// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

// Manager manages secrets and their lifecycle.
type Manager struct {
	provider Provider

	mu    sync.RWMutex
	cache map[string]secretCacheEntry
}

type secretCacheEntry struct {
	value     string
	expiresAt time.Time
}

// NewManager creates a new Secret Manager.
func NewManager(provider Provider) *Manager {
	return &Manager{
		provider: provider,
		cache:    make(map[string]secretCacheEntry),
	}
}

// GetSecret retrieves a secret, checking the cache first.
func (m *Manager) GetSecret(ctx context.Context, id string) (string, error) {
	m.mu.RLock()
	entry, ok := m.cache[id]
	m.mu.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return entry.value, nil
	}

	// Cache miss or expired
	val, err := m.provider.GetSecret(ctx, id)
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	m.cache[id] = secretCacheEntry{
		value:     val,
		expiresAt: time.Now().Add(5 * time.Minute), // Default TTL
	}
	m.mu.Unlock()

	return val, nil
}

// RotateSecret triggers a rotation provided by the underlying provider.
func (m *Manager) RotateSecret(ctx context.Context, id string) (string, error) {
	logging.GetLogger().Info("Rotating secret", "secretID", id)

	newVal, err := m.provider.RotateSecret(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to rotate secret: %w", err)
	}

	// Update cache
	m.mu.Lock()
	m.cache[id] = secretCacheEntry{
		value:     newVal,
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	m.mu.Unlock()

	return newVal, nil
}
