// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package memory provides an in-memory storage implementation for testing.
package memory

import (
	"context"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// ⚡ BOLT: Optimized map key to avoid string allocation/concatenation.
// Randomized Selection from Top 5 High-Impact Targets.
type tokenKey struct {
	userID    string
	serviceID string
}

// Store implements storage.Storage in memory.
//
// Summary: In-memory storage implementation.
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
}

// NewStore creates a new memory store.
//
// Summary: Creates a new in-memory store.
//
// Returns:
//   - *Store: A pointer to the new Store instance.
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
	}
}

// SaveToken saves a user token.
//
// Summary: Persists a user token.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - token: *configv1.UserToken. The token to save.
//
// Returns:
//   - error: Returns nil on success.
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
// Summary: Retrieves a user token.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *configv1.UserToken: The user token if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Removes a user token.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - error: Returns nil on success.
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
// Summary: Loads the complete server configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - *configv1.McpAnyServerConfig: The full server configuration.
//   - error: Returns nil on success.
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
		// ⚡ Bolt Optimization: Append directly to gs.ProfileDefinitions using accessors
		// This avoids creating a temporary object and the overhead of proto.Merge.
		// Randomized Selection from Top 5 High-Impact Targets
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
// Summary: Persists a service configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - service: *configv1.UpstreamServiceConfig. The service configuration to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves a single upstream service configuration by name.
//
// Summary: Retrieves a service configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the service to retrieve.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The service configuration if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Lists all service configurations.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice of all service configurations.
//   - error: Returns nil on success.
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
// Summary: Removes a service configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the service to delete.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close closes the underlying storage connection.
//
// Summary: Closes the store.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
// Memory store always has data available.
//
// Summary: Checks for configuration sources.
//
// Returns:
//   - bool: Always returns true for memory store.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global configuration.
//
// Summary: Retrieves global settings.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - *configv1.GlobalSettings: The global settings.
//   - error: Returns nil on success.
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
// Summary: Persists global settings.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - settings: *configv1.GlobalSettings. The global settings to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
//
// Summary: Lists all secrets.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.Secret: A slice of all secrets.
//   - error: Returns nil on success.
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
// Summary: Retrieves a secret by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the secret.
//
// Returns:
//   - *configv1.Secret: The secret if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Persists a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - secret: *configv1.Secret. The secret to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// Summary: Removes a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the secret.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
//
// Summary: Creates a new user.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - user: *configv1.User. The user to create.
//
// Returns:
//   - error: Returns an error if the user already exists or ID is missing.
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
// Summary: Retrieves a user by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the user.
//
// Returns:
//   - *configv1.User: The user if found, or nil.
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.User: A slice of all users.
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - user: *configv1.User. The user to update.
//
// Returns:
//   - error: Returns an error if the user is not found.
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
// Summary: Removes a user.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the user.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Summary: Lists all profiles.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.ProfileDefinition: A slice of all profiles.
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the profile.
//
// Returns:
//   - *configv1.ProfileDefinition: The profile if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Persists a profile.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - profile: *configv1.ProfileDefinition. The profile to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile deletes a profile definition by name.
//
// Summary: Removes a profile.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the profile.
//
// Returns:
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.Collection: A slice of all service collections.
//   - error: Returns nil on success.
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
// Summary: Retrieves a service collection by name.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the collection.
//
// Returns:
//   - *configv1.Collection: The collection if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Persists a service collection.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - collection: *configv1.Collection. The collection to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
//
// Summary: Removes a service collection.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - name: string. The name of the collection.
//
// Returns:
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//
// Returns:
//   - []*configv1.Credential: A slice of all credentials.
//   - error: Returns nil on success.
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
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the credential.
//
// Returns:
//   - *configv1.Credential: The credential if found, or nil.
//   - error: Returns nil on success.
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
// Summary: Persists a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - cred: *configv1.Credential. The credential to save.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential deletes a credential by ID.
//
// Summary: Removes a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the request (unused in memory store).
//   - id: string. The unique identifier of the credential.
//
// Returns:
//   - error: Returns nil on success.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
