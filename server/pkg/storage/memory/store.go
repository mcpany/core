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
	userID    string
	serviceID string
}

// Store implements storage.Storage in memory.
//
// Summary: A thread-safe, in-memory implementation of the Storage interface, primarily for testing.
type Store struct {
	mu                 sync.RWMutex
	services           map[string]*configv1.UpstreamServiceConfig
	secrets            map[string]*configv1.Secret
	users              map[string]*configv1.User
	profileDefinitions map[string]*configv1.ProfileDefinition
	serviceCollections map[string]*configv1.Collection
	globalSettings     *configv1.GlobalSettings
	tokens             map[tokenKey]*configv1.UserToken
	credentials        map[string]*configv1.Credential
	serviceTemplates   map[string]*configv1.ServiceTemplate
	logs               []*logging.LogEntry
}

// NewStore creates a new memory store.
//
// Summary: Initializes a new, empty in-memory store.
//
// Returns:
//   - *Store: A pointer to the initialized Store.
//
// Side Effects:
//   - Allocates internal maps and slices.
func NewStore() *Store {
	return &Store{
		services:           make(map[string]*configv1.UpstreamServiceConfig),
		secrets:            make(map[string]*configv1.Secret),
		users:              make(map[string]*configv1.User),
		profileDefinitions: make(map[string]*configv1.ProfileDefinition),
		serviceCollections: make(map[string]*configv1.Collection),
		tokens:             make(map[tokenKey]*configv1.UserToken),
		credentials:        make(map[string]*configv1.Credential),
		serviceTemplates:   make(map[string]*configv1.ServiceTemplate),
		logs:               make([]*logging.LogEntry, 0),
	}
}

// SaveLog saves a log entry.
//
// Summary: Appends a log entry to the in-memory log store.
//
// Parameters:
//   - _: context.Context. Unused.
//   - entry: *logging.LogEntry. The log entry to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Appends to the internal logs slice.
func (s *Store) SaveLog(_ context.Context, entry *logging.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, entry)
	return nil
}

// GetRecentLogs retrieves recent log entries.
//
// Summary: Returns the N most recent log entries.
//
// Parameters:
//   - _: context.Context. Unused.
//   - limit: int. The maximum number of logs to return.
//
// Returns:
//   - []*logging.LogEntry: A slice of log entries.
//   - error: Always nil.
//
// Side Effects:
//   - Reads from the internal logs slice.
func (s *Store) GetRecentLogs(_ context.Context, limit int) ([]*logging.LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := len(s.logs)
	if count == 0 {
		return []*logging.LogEntry{}, nil
	}
	start := count - limit
	if start < 0 {
		start = 0
	}
	result := make([]*logging.LogEntry, count-start)
	copy(result, s.logs[start:])
	return result, nil
}

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
// Side Effects:
//   - Updates the internal tokens map.
func (s *Store) SaveToken(_ context.Context, token *configv1.UserToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := tokenKey{
		userID:    token.GetUserId(),
		serviceID: token.GetServiceId(),
	}
	s.tokens[key] = proto.Clone(token).(*configv1.UserToken)
	return nil
}

// GetToken retrieves a user token by user ID and service ID.
//
// Summary: Retrieves a stored user token.
//
// Parameters:
//   - _: context.Context. Unused.
//   - userID: string. The user ID.
//   - serviceID: string. The service ID.
//
// Returns:
//   - *configv1.UserToken: The retrieved token, or nil if not found.
//   - error: Always nil.
func (s *Store) GetToken(_ context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := tokenKey{
		userID:    userID,
		serviceID: serviceID,
	}
	if token, ok := s.tokens[key]; ok {
		return proto.Clone(token).(*configv1.UserToken), nil
	}
	return nil, nil
}

// DeleteToken deletes a user token.
//
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
func (s *Store) DeleteToken(_ context.Context, userID, serviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := tokenKey{
		userID:    userID,
		serviceID: serviceID,
	}
	delete(s.tokens, key)
	return nil
}

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
			current = append(current, proto.Clone(p).(*configv1.ProfileDefinition))
		}
		gs.SetProfileDefinitions(current)
	}

	// 4. Build final ServerConfig
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: upstreamServices,
		GlobalSettings:   gs,
	}.Build()

	return cfg, nil
}

// SaveService saves a single upstream service configuration.
//
// Summary: Stores an upstream service configuration.
//
// Parameters:
//   - _: context.Context. Unused.
//   - service: *configv1.UpstreamServiceConfig. The service config to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal services map.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves a single upstream service configuration by name.
//
// Summary: Retrieves an upstream service configuration.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The name of the service.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The service config, or nil if not found.
//   - error: Always nil.
func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

// ListServices lists all upstream service configurations.
//
// Summary: Lists all stored upstream service configurations.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of service configs.
//   - error: Always nil.
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
//
// Summary: Deletes an upstream service configuration.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The name of the service to delete.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal services map.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close closes the underlying storage connection.
//
// Summary: No-op for in-memory store.
//
// Returns:
//   - error: Always nil.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
//
// Summary: Indicates if the store supports config sources (always true for this mock).
//
// Returns:
//   - bool: Always true.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global configuration.
//
// Summary: Retrieves the global settings object.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - *configv1.GlobalSettings: The global settings.
//   - error: Always nil.
func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// SaveGlobalSettings saves the global configuration.
//
// Summary: Persists the global settings.
//
// Parameters:
//   - _: context.Context. Unused.
//   - settings: *configv1.GlobalSettings. The settings to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal global settings.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
//
// Summary: Lists all stored secrets.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.Secret: A list of secrets.
//   - error: Always nil.
func (s *Store) ListSecrets(_ context.Context) ([]*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Secret, 0, len(s.secrets))
	for _, secret := range s.secrets {
		list = append(list, proto.Clone(secret).(*configv1.Secret))
	}
	return list, nil
}

// GetSecret retrieves a secret by ID.
//
// Summary: Retrieves a secret by its ID.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The secret ID.
//
// Returns:
//   - *configv1.Secret: The secret, or nil if not found.
//   - error: Always nil.
func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// SaveSecret saves a secret.
//
// Summary: Stores a secret.
//
// Parameters:
//   - _: context.Context. Unused.
//   - secret: *configv1.Secret. The secret to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal secrets map.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// Summary: Deletes a secret.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The secret ID.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal secrets map.
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
	s.mu.Lock()
	defer s.mu.Unlock()
	if user.GetId() == "" {
		return fmt.Errorf("user ID is required")
	}
	if _, ok := s.users[user.GetId()]; ok {
		return fmt.Errorf("user already exists")
	}
	s.users[user.GetId()] = proto.Clone(user).(*configv1.User)
	return nil
}

// GetUser retrieves a user by ID.
//
// Summary: Retrieves a user.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The user ID.
//
// Returns:
//   - *configv1.User: The user, or nil if not found.
//   - error: Always nil.
func (s *Store) GetUser(_ context.Context, id string) (*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[id]; ok {
		return proto.Clone(user).(*configv1.User), nil
	}
	return nil, nil
}

// ListUsers retrieves all users.
//
// Summary: Lists all users.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.User: A list of users.
//   - error: Always nil.
func (s *Store) ListUsers(_ context.Context) ([]*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.User, 0, len(s.users))
	for _, user := range s.users {
		list = append(list, proto.Clone(user).(*configv1.User))
	}
	return list, nil
}

// UpdateUser updates an existing user.
//
// Summary: Updates an existing user.
//
// Parameters:
//   - _: context.Context. Unused.
//   - user: *configv1.User. The user to update.
//
// Returns:
//   - error: An error if the user is not found.
//
// Errors:
//   - Returns "user not found" if the user does not exist.
//
// Side Effects:
//   - Updates the internal users map.
func (s *Store) UpdateUser(_ context.Context, user *configv1.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.GetId()]; !ok {
		return fmt.Errorf("user not found")
	}
	s.users[user.GetId()] = proto.Clone(user).(*configv1.User)
	return nil
}

// DeleteUser deletes a user by ID.
//
// Summary: Deletes a user.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The user ID.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal users map.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Summary: Lists all stored profile definitions.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.ProfileDefinition: A list of profiles.
//   - error: Always nil.
func (s *Store) ListProfiles(_ context.Context) ([]*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.ProfileDefinition, 0, len(s.profileDefinitions))
	for _, p := range s.profileDefinitions {
		list = append(list, proto.Clone(p).(*configv1.ProfileDefinition))
	}
	return list, nil
}

// GetProfile retrieves a profile definition by name.
//
// Summary: Retrieves a profile by name.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The profile name.
//
// Returns:
//   - *configv1.ProfileDefinition: The profile, or nil if not found.
//   - error: Always nil.
func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

// SaveProfile saves a profile definition.
//
// Summary: Stores a profile definition.
//
// Parameters:
//   - _: context.Context. Unused.
//   - profile: *configv1.ProfileDefinition. The profile to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal profile map.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile deletes a profile definition by name.
//
// Summary: Deletes a profile.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The profile name.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal profile map.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
//
// Summary: Lists all service collections.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.Collection: A list of collections.
//   - error: Always nil.
func (s *Store) ListServiceCollections(_ context.Context) ([]*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Collection, 0, len(s.serviceCollections))
	for _, c := range s.serviceCollections {
		list = append(list, proto.Clone(c).(*configv1.Collection))
	}
	return list, nil
}

// GetServiceCollection retrieves a service collection by name.
//
// Summary: Retrieves a service collection.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The collection name.
//
// Returns:
//   - *configv1.Collection: The collection, or nil if not found.
//   - error: Always nil.
func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.Collection), nil
	}
	return nil, nil
}

// SaveServiceCollection saves a service collection.
//
// Summary: Stores a service collection.
//
// Parameters:
//   - _: context.Context. Unused.
//   - collection: *configv1.Collection. The collection to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal collection map.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
//
// Summary: Deletes a service collection.
//
// Parameters:
//   - _: context.Context. Unused.
//   - name: string. The collection name.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal collection map.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// ListCredentials retrieves all credentials.
//
// Summary: Lists all credentials.
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - []*configv1.Credential: A list of credentials.
//   - error: Always nil.
func (s *Store) ListCredentials(_ context.Context) ([]*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Credential, 0, len(s.credentials))
	for _, c := range s.credentials {
		list = append(list, proto.Clone(c).(*configv1.Credential))
	}
	return list, nil
}

// GetCredential retrieves a credential by ID.
//
// Summary: Retrieves a credential by ID.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The credential ID.
//
// Returns:
//   - *configv1.Credential: The credential, or nil if not found.
//   - error: Always nil.
func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

// SaveCredential saves a credential.
//
// Summary: Stores a credential.
//
// Parameters:
//   - _: context.Context. Unused.
//   - cred: *configv1.Credential. The credential to save.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Updates the internal credential map.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential deletes a credential by ID.
//
// Summary: Deletes a credential.
//
// Parameters:
//   - _: context.Context. Unused.
//   - id: string. The credential ID.
//
// Returns:
//   - error: Always nil.
//
// Side Effects:
//   - Removes from the internal credential map.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
