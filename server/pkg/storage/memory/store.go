// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package memory provides an in-memory storage implementation for testing.
package memory

import (
	"context"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/proto"
)

// tokenKey represents a composite key for storing user tokens.
//
// Summary: Composite key structure for indexing user tokens by user and service.
type tokenKey struct {
// Store implements storage.Storage in memory.
//
// Summary: A thread-safe, in-memory implementation of the Storage interface, primarily for testing.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Store struct {
	mu                 sync.RWMutex
	services           map[string]*configv1.UpstreamServiceConfig
	secrets            map[string]*configv1.Secret
// NewStore creates a new memory store.
//
// Summary: Initializes a new, empty in-memory store.
//
// Returns:
//   - *Store: A pointer to the initialized Store.
//
// Side Effects:
//   - Allocates internal maps and slices.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// SaveLog saves a log entry.
//
// Summary: Appends a log entry to the in-memory log store.
//
// Parameters:
// GetRecentLogs retrieves recent log entries.
//
// Summary: Returns the N most recent log entries.
//
// Parameters:
//   - _: context.Context. Unused.
//   - limit: int. The maximum number of logs to return.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - []*logging.LogEntry: A slice of log entries.
//   - error: Always nil.
//
// Side Effects:
//   - Reads from the internal logs slice.
// Errors:
//   - triggers relevant error states on failure.
func (s *Store) GetRecentLogs(_ context.Context, limit int) ([]*logging.LogEntry, error) {
	s.mu.RLock()
// SaveToken saves a user token.
//
// Summary: Stores a user token in memory.
//
// Parameters:
//   - _: context.Context. Unused.
//   - token: *configv1.UserToken. The token to save.
//
// Returns:
//   - error: Always nil.
//
// GetToken retrieves a user token by user ID and service ID.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a stored user token.
//
// Parameters:
//   - _: context.Context. Unused.
//   - userID: string. The user ID.
//   - serviceID: string. The service ID.
//
// Returns:
// DeleteToken deletes a user token.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Removes a user token from memory.
//
// Parameters:
//   - _: context.Context. Unused.
//   - userID: string. The user ID.
//   - serviceID: string. The service ID.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Deletes from the internal tokens map.
// Errors:
//   - triggers relevant error states on failure.
// Load retrieves the full server configuration.
//
// Summary: Constructs and returns the complete server configuration from stored components.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The complete configuration object.
//   - error: Always nil.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (s *Store) Load(_ context.Context) (*configv1.McpAnyServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Collect Upstream Services
	upstreamServices := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		upstreamServices = append(upstreamServices, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}

	// 2. Prepare Global Settings
	var gs *configv1.GlobalSettings
	if s.globalSettings != nil {
		gs = proto.Clone(s.globalSettings).(*configv1.GlobalSettings)
	} else {
		// Start empty if nil, but if we have profiles we need a base object.
		// Builder{}.Build() returns a valid opaque object (pointer).
		gs = configv1.GlobalSettings_builder{}.Build()
	}

	// 3. Merge Profiles into Global Settings
	if len(s.profileDefinitions) > 0 {
		current := gs.GetProfileDefinitions()
		for _, p := range s.profileDefinitions {
// SaveService saves a single upstream service configuration.
//
// Summary: Stores an upstream service configuration.
//
// Parameters:
//   - _: context.Context. Unused.
//   - service: *configv1.UpstreamServiceConfig. The service config to save.
//
// GetService retrieves a single upstream service configuration by name.
//
// Summary: Retrieves an upstream service configuration.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
//   - name: string. The name of the service.
//
// Returns:
// ListServices lists all upstream service configurations.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Lists all stored upstream service configurations.
//
// Parameters:
//   - _: context.Context. Unused.
// DeleteService deletes an upstream service configuration by name.
//
// Summary: Deletes an upstream service configuration.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The name of the service to delete.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal services map.
// Errors:
//   - triggers relevant error states on failure.
// Close closes the underlying storage connection.
//
// GetGlobalSettings retrieves the global configuration.
//
// Summary: Retrieves the global settings object.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
// SaveGlobalSettings saves the global configuration.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Persists the global settings.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
//   - settings: *configv1.GlobalSettings. The settings to save.
//
// Returns:
// ListSecrets retrieves all secrets.
//
// Summary: Lists all stored secrets.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
// GetSecret retrieves a secret by ID.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a secret by its ID.
//
// Parameters:
//   - _: context.Context. Unused.
// SaveSecret saves a secret.
//
// Summary: Stores a secret.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
// DeleteSecret deletes a secret by ID.
//
// Summary: Deletes a secret.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The secret ID.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal secrets map.
// Errors:
//   - triggers relevant error states on failure.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
//
// Summary: Creates a new user entry.
//
// Parameters:
//   - _: context.Context. Unused.
//   - user: *configv1.User. The user to create.
//
// Returns:
//   - error: An error if the user ID is missing or already exists.
//
// Errors:
//   - Returns "user ID is required" if ID is empty.
//   - Returns "user already exists" if ID is present.
//
// Side Effects:
//   - Adds to the internal users map.
func (s *Store) CreateUser(_ context.Context, user *configv1.User) error {
// GetUser retrieves a user by ID.
//
// Summary: Retrieves a user.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The user ID.
//
// Returns:
// ListUsers retrieves all users.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Lists all users.
// UpdateUser updates an existing user.
//
// Summary: Updates an existing user.
//
// Parameters:
//   - _: context.Context. Unused.
//   - user: *configv1.User. The user to update.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: An error if the user is not found.
//
// Errors:
// DeleteUser deletes a user by ID.
//
// Summary: Deletes a user.
//
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The user ID.
//
// Profiles
// ListProfiles retrieves all profile definitions.
//
// Summary: Lists all stored profile definitions.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
// GetProfile retrieves a profile definition by name.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a profile by name.
//
// Parameters:
//   - _: context.Context. Unused.
// SaveProfile saves a profile definition.
//
// Summary: Stores a profile definition.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
// DeleteProfile deletes a profile definition by name.
//
// Summary: Deletes a profile.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The profile name.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Service Collections
// ListServiceCollections retrieves all service collections.
//
// Summary: Lists all service collections.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
// GetServiceCollection retrieves a service collection by name.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a service collection.
//
// Parameters:
//   - _: context.Context. Unused.
// SaveServiceCollection saves a service collection.
//
// Summary: Stores a service collection.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
// DeleteServiceCollection deletes a service collection by name.
//
// Summary: Deletes a service collection.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The collection name.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Credentials
// ListCredentials retrieves all credentials.
//
// Summary: Lists all credentials.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
// GetCredential retrieves a credential by ID.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Retrieves a credential by ID.
//
// Parameters:
//   - _: context.Context. Unused.
// SaveCredential saves a credential.
//
// Summary: Stores a credential.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - _: context.Context. Unused.
// DeleteCredential deletes a credential by ID.
//
// Summary: Deletes a credential.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The credential ID.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal credential map.
// Errors:
//   - triggers relevant error states on failure.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
