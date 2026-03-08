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

// Manager handles the lifecycle and resolution of profiles. Summary: Manages profile definitions and resolution.
//
// Summary: Manager handles the lifecycle and resolution of profiles. Summary: Manages profile definitions and resolution.
//
// Fields:
//   - Contains the configuration and state properties required for Manager functionality.
type Manager struct {
	mu       sync.RWMutex
	profiles map[string]*configv1.ProfileDefinition
}

// NewManager creates a new Profile Manager. Summary: Initializes a new Profile Manager. Parameters: - profiles: []*configv1.ProfileDefinition. Initial profile definitions. Returns: - *Manager: The initialized manager.
//
// Summary: NewManager creates a new Profile Manager. Summary: Initializes a new Profile Manager. Parameters: - profiles: []*configv1.ProfileDefinition. Initial profile definitions. Returns: - *Manager: The initialized manager.
//
// Parameters:
//   - profiles ([]*configv1.ProfileDefinition): The profiles parameter used in the operation.
//
// Returns:
//   - (*Manager): The resulting Manager object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewManager(profiles []*configv1.ProfileDefinition) *Manager {
	m := &Manager{
		profiles: make(map[string]*configv1.ProfileDefinition),
	}
	m.Update(profiles)
	return m
}

// Update updates the profile definitions managed by the manager. Summary: Updates the stored profile definitions. Parameters: - profiles: []*configv1.ProfileDefinition. The new list of profiles.
//
// Summary: Update updates the profile definitions managed by the manager. Summary: Updates the stored profile definitions. Parameters: - profiles: []*configv1.ProfileDefinition. The new list of profiles.
//
// Parameters:
//   - profiles ([]*configv1.ProfileDefinition): The profiles parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
func (m *Manager) Update(profiles []*configv1.ProfileDefinition) {
	newProfiles := make(map[string]*configv1.ProfileDefinition)
	for _, p := range profiles {
		newProfiles[p.GetName()] = p
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.profiles = newProfiles
}

// GetProfileDefinition returns the profile definition by name. Summary: Retrieves a profile definition. Parameters: - name: string. The name of the profile. Returns: - *configv1.ProfileDefinition: The profile definition. - bool: True if found.
//
// Summary: GetProfileDefinition returns the profile definition by name. Summary: Retrieves a profile definition. Parameters: - name: string. The name of the profile. Returns: - *configv1.ProfileDefinition: The profile definition. - bool: True if found.
//
// Parameters:
//   - name (string): The name parameter used in the operation.
//
// Returns:
//   - (*configv1.ProfileDefinition): The resulting configv1.ProfileDefinition object containing the requested data.
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *Manager) GetProfileDefinition(name string) (*configv1.ProfileDefinition, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.profiles[name]
	return p, ok
}

// ResolveProfile computes the final effective configuration for a given profile, applying inheritance and overrides. Summary: Resolves a profile hierarchy into a final configuration. Parameters: - profileName: string. The name of the profile to resolve. Returns: - map[string]*configv1.ProfileServiceConfig: Merged service configs. - map[string]*configv1.SecretValue: Merged secrets. - error: Error if profile not found or cycle detected.
//
// Summary: ResolveProfile computes the final effective configuration for a given profile, applying inheritance and overrides. Summary: Resolves a profile hierarchy into a final configuration. Parameters: - profileName: string. The name of the profile to resolve. Returns: - map[string]*configv1.ProfileServiceConfig: Merged service configs. - map[string]*configv1.SecretValue: Merged secrets. - error: Error if profile not found or cycle detected.
//
// Parameters:
//   - profileName (string): The profile name parameter used in the operation.
//
// Returns:
//   - (map[string]*configv1.ProfileServiceConfig): A string value representing the operation's result.
//   - (map[string]*configv1.SecretValue): A string value representing the operation's result.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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
