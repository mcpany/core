// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package memory provides an in-memory storage implementation for testing.
package memory

import (
	"context"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Store implements storage.Storage in memory.
type Store struct {
	mu       sync.RWMutex
	services map[string]*configv1.UpstreamServiceConfig
}

// NewStore creates a new memory store.
func NewStore() *Store {
	return &Store{
		services: make(map[string]*configv1.UpstreamServiceConfig),
	}
}

// Load retrieves the full server configuration.
func (s *Store) Load(_ context.Context) (*configv1.McpAnyServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := &configv1.McpAnyServerConfig{}
	// Pre-allocate slice if we knew the size, but here we iterate map.
	// We can count map keys.
	cfg.UpstreamServices = make([]*configv1.UpstreamServiceConfig, 0, len(s.services))

	for _, svc := range s.services {
		cfg.UpstreamServices = append(cfg.UpstreamServices, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}
	return cfg, nil
}

// SaveService saves a single upstream service configuration.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves a single upstream service configuration by name.
func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

// ListServices lists all upstream service configurations.
func (s *Store) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		list = append(list, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}
	return list, nil
}

// DeleteService deletes an upstream service configuration by name.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close closes the underlying storage connection.
func (s *Store) Close() error {
	return nil
}
