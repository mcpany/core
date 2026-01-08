// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// MemoryStore implements the Store interface for in-memory configuration.
// It allows for dynamic updates to the configuration without file changes.
type MemoryStore struct {
	mu     sync.RWMutex
	config *configv1.McpAnyServerConfig
}

// NewMemoryStore creates a new, empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Load returns the current configuration stored in memory.
// It returns a clone of the configuration to ensure thread safety.
func (s *MemoryStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil {
		return nil, nil
	}

	return proto.Clone(s.config).(*configv1.McpAnyServerConfig), nil
}

// Update replaces the stored configuration with the provided one.
func (s *MemoryStore) Update(cfg *configv1.McpAnyServerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = proto.Clone(cfg).(*configv1.McpAnyServerConfig)
}
