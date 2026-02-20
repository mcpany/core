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

// NewStore initializes a new storage backend.
//
// Summary: Creates and returns a new Store instance.
//
// Returns:
//   - *Store: The initialized store.
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

// SaveToken persists a user token.
//
// Summary: Saves or updates a user token.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - token: *configv1.UserToken. The token to save.
//
// Returns:
//   - error: An error if the operation fails.
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

// GetToken retrieves a user token.
//
// Summary: Fetches a user token by user ID and service ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *configv1.UserToken: The user token, or nil if not found.
//   - error: An error if the operation fails.
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

// DeleteToken removes a user token.
//
// Summary: Deletes a user token by user ID and service ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - error: An error if the operation fails.
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
// Summary: Loads and aggregates all configuration entities.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The complete server configuration.
//   - error: An error if loading fails.
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

// SaveService persists an upstream service configuration.
//
// Summary: Saves or updates an upstream service.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - service: *configv1.UpstreamServiceConfig. The service configuration to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves an upstream service configuration.
//
// Summary: Fetches a service configuration by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The unique name of the service.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The service configuration, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

// ListServices retrieves all upstream service configurations.
//
// Summary: Lists all registered upstream services.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of service configurations.
//   - error: An error if the operation fails.
func (s *Store) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		list = append(list, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}
	return list, nil
}

// DeleteService removes an upstream service configuration.
//
// Summary: Deletes a service configuration by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The unique name of the service.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close terminates the storage connection.
//
// Summary: Closes the underlying storage connection.
//
// Returns:
//   - error: An error if the closure fails.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources checks if the store has external configuration sources.
//
// Summary: Indicates if the store relies on external config files.
//
// Returns:
//   - bool: True if external sources are used, false otherwise.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global server settings.
//
// Summary: Fetches the global configuration settings.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - *configv1.GlobalSettings: The global settings.
//   - error: An error if the operation fails.
func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// SaveGlobalSettings persists the global server settings.
//
// Summary: Saves or updates the global configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - settings: *configv1.GlobalSettings. The settings to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all stored secrets.
//
// Summary: Lists all secrets.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.Secret: A list of secrets.
//   - error: An error if the operation fails.
func (s *Store) ListSecrets(_ context.Context) ([]*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Secret, 0, len(s.secrets))
	for _, secret := range s.secrets {
		list = append(list, proto.Clone(secret).(*configv1.Secret))
	}
	return list, nil
}

// GetSecret retrieves a secret by its ID.
//
// Summary: Fetches a secret by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the secret.
//
// Returns:
//   - *configv1.Secret: The secret, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// SaveSecret persists a secret.
//
// Summary: Saves or updates a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - secret: *configv1.Secret. The secret to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret removes a secret.
//
// Summary: Deletes a secret by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the secret.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
//
// Summary: Creates a new user record.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - user: *configv1.User. The user to create.
//
// Returns:
//   - error: An error if the user already exists or operation fails.
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
// Summary: Fetches a user by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the user.
//
// Returns:
//   - *configv1.User: The user, or nil if not found.
//   - error: An error if the operation fails.
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
// Summary: Lists all registered users.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.User: A list of users.
//   - error: An error if the operation fails.
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
// Summary: Updates an existing user record.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - user: *configv1.User. The user to update.
//
// Returns:
//   - error: An error if the user does not exist or operation fails.
func (s *Store) UpdateUser(_ context.Context, user *configv1.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.GetId()]; !ok {
		return fmt.Errorf("user not found")
	}
	s.users[user.GetId()] = proto.Clone(user).(*configv1.User)
	return nil
}

// DeleteUser removes a user.
//
// Summary: Deletes a user by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the user.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Summary: Lists all profile definitions.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.ProfileDefinition: A list of profile definitions.
//   - error: An error if the operation fails.
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
// Summary: Fetches a profile definition by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the profile.
//
// Returns:
//   - *configv1.ProfileDefinition: The profile definition, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

// SaveProfile persists a profile definition.
//
// Summary: Saves or updates a profile definition.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - profile: *configv1.ProfileDefinition. The profile to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile removes a profile definition.
//
// Summary: Deletes a profile definition by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the profile.
//
// Returns:
//   - error: An error if the operation fails.
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
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.Collection: A list of service collections.
//   - error: An error if the operation fails.
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
// Summary: Fetches a service collection by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the collection.
//
// Returns:
//   - *configv1.Collection: The service collection, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.Collection), nil
	}
	return nil, nil
}

// SaveServiceCollection persists a service collection.
//
// Summary: Saves or updates a service collection.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - collection: *configv1.Collection. The collection to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection removes a service collection.
//
// Summary: Deletes a service collection by name.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - name: string. The name of the collection.
//
// Returns:
//   - error: An error if the operation fails.
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
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - []*configv1.Credential: A list of credentials.
//   - error: An error if the operation fails.
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
// Summary: Fetches a credential by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the credential.
//
// Returns:
//   - *configv1.Credential: The credential, or nil if not found.
//   - error: An error if the operation fails.
func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

// SaveCredential persists a credential.
//
// Summary: Saves or updates a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - cred: *configv1.Credential. The credential to save.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential removes a credential.
//
// Summary: Deletes a credential by ID.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - id: string. The unique identifier of the credential.
//
// Returns:
//   - error: An error if the operation fails.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
