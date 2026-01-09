// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package profile provides functionality for managing and resolving profiles.
package profile

import (
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Manager handles the lifecycle and resolution of profiles.
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*configv1.ProfileDefinition
}

// NewManager creates a new Profile Manager.
func NewManager(profiles []*configv1.ProfileDefinition) *Manager {
	m := &Manager{
		profiles: make(map[string]*configv1.ProfileDefinition),
	}
	m.Update(profiles)
	return m
}

// Update updates the profile definitions managed by the manager.
func (m *Manager) Update(profiles []*configv1.ProfileDefinition) {
	newProfiles := make(map[string]*configv1.ProfileDefinition)
	for _, p := range profiles {
		newProfiles[p.GetName()] = p
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles = newProfiles
}

// GetProfileDefinition returns the profile definition by name.
func (m *Manager) GetProfileDefinition(name string) (*configv1.ProfileDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[name]
	return p, ok
}

// ResolveProfile computes the final effective configuration for a given profile,
// applying inheritance and overrides.
// It returns a map of ProfileServiceConfigs and a map of resolved Secrets.
func (m *Manager) ResolveProfile(profileName string) (map[string]*configv1.ProfileServiceConfig, map[string]*configv1.SecretValue, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 1. Find the profile
	_, ok := m.profiles[profileName]
	if !ok {
		return nil, nil, fmt.Errorf("profile not found: %s", profileName)
	}

	// 2. Collect hierarchy (DFS or simple list if we assume directed acyclic graph)
	// We want to apply parents FIRST, then children.
	// visited check to prevent cycles.
	visited := make(map[string]bool)
	chain := []*configv1.ProfileDefinition{}

	var collect func(name string) error
	collect = func(name string) error {
		if visited[name] {
			return fmt.Errorf("cycle detected in profile inheritance: %s", name)
		}
		visited[name] = true
		p, exists := m.profiles[name]
		if !exists {
			return fmt.Errorf("parent profile not found: %s", name)
		}

		// Process parents first (so they are earlier in the chain, wait...)
		// If we want Parent settings to be defaults, and Child to override,
		// we should apply Parent *then* Child.
		// So we recurse first, then append Self.
		for _, parentID := range p.GetParentProfileIds() {
			if err := collect(parentID); err != nil {
				return err
			}
		}
		chain = append(chain, p)
		return nil
	}

	if err := collect(profileName); err != nil {
		return nil, nil, err
	}

	// 3. Merge
	// Resulting overrides
	mergedConfigs := make(map[string]*configv1.ProfileServiceConfig)
	mergedSecrets := make(map[string]*configv1.SecretValue)

	for _, p := range chain {
		// Merge Service Configs
		for serviceID, config := range p.GetServiceConfig() {
			// Deep copy to avoid mutating original
			newConfig := proto.Clone(config).(*configv1.ProfileServiceConfig)

			existing, exists := mergedConfigs[serviceID]
			if exists {
				// Merge existing (parent) with new (child)
				proto.Merge(existing, newConfig)
			} else {
				mergedConfigs[serviceID] = newConfig
			}
		}

		// Merge Secrets
		for key, secret := range p.GetSecrets() {
			// Deep copy
			newSecret := proto.Clone(secret).(*configv1.SecretValue)
			// Child overrides parent completely for the same key
			mergedSecrets[key] = newSecret
		}
	}

	return mergedConfigs, mergedSecrets, nil
}
