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
// Summary: implements storage.Storage in memory.
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
}

// NewStore creates a new memory store.
//
// Summary: creates a new memory store.
//
// Parameters:
//   None.
//
// Returns:
//   - *Store: The *Store.
func NewStore() *Store {
	return &Store{
		services:           make(map[string]*configv1.UpstreamServiceConfig),
		secrets:            make(map[string]*configv1.Secret),
		users:              make(map[string]*configv1.User),
		profileDefinitions: make(map[string]*configv1.ProfileDefinition),
		serviceCollections: make(map[string]*configv1.Collection),
		tokens:             make(map[tokenKey]*configv1.UserToken),
		credentials:        make(map[string]*configv1.Credential),
	}
}

// SaveToken saves a user token.
//
// Summary: saves a user token.
//
// Parameters:
//   - _: context.Context. The _.
//   - token: *configv1.UserToken. The token.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a user token by user ID and service ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - userID: string. The userID.
//   - serviceID: string. The serviceID.
//
// Returns:
//   - *configv1.UserToken: The *configv1.UserToken.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: deletes a user token.
//
// Parameters:
//   - _: context.Context. The _.
//   - userID: string. The userID.
//   - serviceID: string. The serviceID.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves the full server configuration.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The *configv1.McpAnyServerConfig.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves a single upstream service configuration.
//
// Parameters:
//   - _: context.Context. The _.
//   - service: *configv1.UpstreamServiceConfig. The service.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves a single upstream service configuration by name.
//
// Summary: retrieves a single upstream service configuration by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The *configv1.UpstreamServiceConfig.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: lists all upstream service configurations.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The []*configv1.UpstreamServiceConfig.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: deletes an upstream service configuration by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close closes the underlying storage connection.
//
// Summary: closes the underlying storage connection.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
//
// Summary: returns true if the store has configuration sources (e.g., file paths) configured.
//
// Parameters:
//   None.
//
// Returns:
//   - bool: The bool.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global configuration.
//
// Summary: retrieves the global configuration.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - *configv1.GlobalSettings: The *configv1.GlobalSettings.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves the global configuration.
//
// Parameters:
//   - _: context.Context. The _.
//   - settings: *configv1.GlobalSettings. The settings.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
//
// Summary: retrieves all secrets.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.Secret: The []*configv1.Secret.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a secret by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - *configv1.Secret: The *configv1.Secret.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves a secret.
//
// Parameters:
//   - _: context.Context. The _.
//   - secret: *configv1.Secret. The secret.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// Summary: deletes a secret by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
//
// Summary: creates a new user.
//
// Parameters:
//   - _: context.Context. The _.
//   - user: *configv1.User. The user.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a user by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - *configv1.User: The *configv1.User.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves all users.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.User: The []*configv1.User.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: updates an existing user.
//
// Parameters:
//   - _: context.Context. The _.
//   - user: *configv1.User. The user.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: deletes a user by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Summary: retrieves all profile definitions.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.ProfileDefinition: The []*configv1.ProfileDefinition.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a profile definition by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - *configv1.ProfileDefinition: The *configv1.ProfileDefinition.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves a profile definition.
//
// Parameters:
//   - _: context.Context. The _.
//   - profile: *configv1.ProfileDefinition. The profile.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile deletes a profile definition by name.
//
// Summary: deletes a profile definition by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
//
// Summary: retrieves all service collections.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.Collection: The []*configv1.Collection.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a service collection by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - *configv1.Collection: The *configv1.Collection.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves a service collection.
//
// Parameters:
//   - _: context.Context. The _.
//   - collection: *configv1.Collection. The collection.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
//
// Summary: deletes a service collection by name.
//
// Parameters:
//   - _: context.Context. The _.
//   - name: string. The name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// ListCredentials retrieves all credentials.
//
// Summary: retrieves all credentials.
//
// Parameters:
//   - _: context.Context. The _.
//
// Returns:
//   - []*configv1.Credential: The []*configv1.Credential.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: retrieves a credential by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - *configv1.Credential: The *configv1.Credential.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: saves a credential.
//
// Parameters:
//   - _: context.Context. The _.
//   - cred: *configv1.Credential. The cred.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential deletes a credential by ID.
//
// Summary: deletes a credential by ID.
//
// Parameters:
//   - _: context.Context. The _.
//   - id: string. The id.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
