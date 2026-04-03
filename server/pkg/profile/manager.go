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

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*configv1.ProfileDefinition
}

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewManager(profiles []*configv1.ProfileDefinition) *Manager {
	m := &Manager{
		profiles: make(map[string]*configv1.ProfileDefinition),
	}
	m.Update(profiles)
	return m
}

// Update serves as a public interface for interacting with Update.
//
// Summary: Update the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) Update(profiles []*configv1.ProfileDefinition) {
	newProfiles := make(map[string]*configv1.ProfileDefinition)
	for _, p := range profiles {
		newProfiles[p.GetName()] = p
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles = newProfiles
}

// GetProfileDefinition serves as a public interface for interacting with GetProfileDefinition.
//
// Summary: Fetches and returns the underlying profile definition from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *Manager) GetProfileDefinition(name string) (*configv1.ProfileDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[name]
	return p, ok
}

// ResolveProfile serves as a public interface for interacting with ResolveProfile.
//
// Summary: Resolve the profile appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
	// We use two maps to track state:
	// - processing: currently in the recursion stack (detects cycles)
	// - visited: fully processed and added to chain (prevents duplicates)
	processing := make(map[string]bool)
	visited := make(map[string]bool)
	chain := []*configv1.ProfileDefinition{}

	var collect func(name string) error
	collect = func(name string) error {
		if processing[name] {
			return fmt.Errorf("cycle detected in profile inheritance: %s", name)
		}
		if visited[name] {
			return nil
		}

		processing[name] = true
		defer func() {
			delete(processing, name)
			visited[name] = true
		}()

		p, exists := m.profiles[name]
		if !exists {
			return fmt.Errorf("parent profile not found: %s", name)
		}

		// Process parents first
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
