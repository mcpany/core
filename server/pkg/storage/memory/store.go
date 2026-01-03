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
	mu             sync.RWMutex
	services       map[string]*configv1.UpstreamServiceConfig
	secrets        map[string]*configv1.Secret
	globalSettings *configv1.GlobalSettings
}

// NewStore creates a new memory store.
func NewStore() *Store {
	return &Store{
		services: make(map[string]*configv1.UpstreamServiceConfig),
		secrets:  make(map[string]*configv1.Secret),
	}
}

// Load retrieves the full server configuration.
func (s *Store) Load(_ context.Context) (*configv1.McpAnyServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}

	builder := configv1.McpAnyServerConfig_builder{
		UpstreamServices: services,
	}

	if s.globalSettings != nil {
		builder.GlobalSettings = proto.Clone(s.globalSettings).(*configv1.GlobalSettings)
	}

	return builder.Build(), nil
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

// GetGlobalSettings retrieves the global configuration.
func (s *Store) GetGlobalSettings() (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// SaveGlobalSettings saves the global configuration.
func (s *Store) SaveGlobalSettings(settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
func (s *Store) ListSecrets() ([]*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Secret, 0, len(s.secrets))
	for _, secret := range s.secrets {
		list = append(list, proto.Clone(secret).(*configv1.Secret))
	}
	return list, nil
}

// GetSecret retrieves a secret by ID.
func (s *Store) GetSecret(id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// SaveSecret saves a secret.
func (s *Store) SaveSecret(secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
func (s *Store) DeleteSecret(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}
