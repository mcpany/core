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
// It is primarily used for testing and ephemeral configurations.
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

// NewStore initializes a new in-memory store.
//
// Returns:
//  *Store: A pointer to the initialized Store instance.
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

// SaveToken saves or updates a user token.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  token (*configv1.UserToken): The user token to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal tokens map.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  userID (string): The unique identifier of the user.
//  serviceID (string): The unique identifier of the service.
//
// Returns:
//  *configv1.UserToken: The user token if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  userID (string): The unique identifier of the user.
//  serviceID (string): The unique identifier of the service.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the token from the internal map.
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

// Load retrieves the full server configuration from the in-memory store.
// It aggregates services, settings, collections, and profiles into a single configuration object.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  *configv1.McpAnyServerConfig: The complete server configuration.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  service (*configv1.UpstreamServiceConfig): The service configuration to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal services map.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService retrieves a single upstream service configuration by name.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the service.
//
// Returns:
//  *configv1.UpstreamServiceConfig: The service configuration if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.UpstreamServiceConfig: A slice of all service configurations.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the service to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the service from the internal map.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close closes the underlying storage connection.
// For memory store, this is a no-op.
//
// Returns:
//  error: Always nil.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources checks if the store has configuration sources configured.
// For memory store, this always returns true.
//
// Returns:
//  bool: Always true.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global configuration.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  *configv1.GlobalSettings: The global settings.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  settings (*configv1.GlobalSettings): The settings to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal global settings.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.Secret: A slice of all secrets.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the secret.
//
// Returns:
//  *configv1.Secret: The secret object if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  secret (*configv1.Secret): The secret object to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal secrets map.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the secret to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the secret from the internal map.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  user (*configv1.User): The user object to create.
//
// Returns:
//  error: An error if the user ID is missing or if the user already exists.
//
// Side Effects:
//  Adds a new user to the internal map.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the user.
//
// Returns:
//  *configv1.User: The user object if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.User: A slice of all user objects.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  user (*configv1.User): The user object with updated information.
//
// Returns:
//  error: An error if the user is not found.
//
// Side Effects:
//  Updates the user in the internal map.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the user to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the user from the internal map.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.ProfileDefinition: A slice of all profile definitions.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the profile.
//
// Returns:
//  *configv1.ProfileDefinition: The profile definition if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  profile (*configv1.ProfileDefinition): The profile definition to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal profile definitions map.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile deletes a profile definition by name.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the profile to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the profile definition from the internal map.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.Collection: A slice of all service collections.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the collection.
//
// Returns:
//  *configv1.Collection: The service collection if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  collection (*configv1.Collection): The service collection to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal service collections map.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  name (string): The name of the collection to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the service collection from the internal map.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// ListCredentials retrieves all credentials.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//
// Returns:
//  []*configv1.Credential: A slice of all credentials.
//  error: Always nil.
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the credential.
//
// Returns:
//  *configv1.Credential: The credential object if found.
//  error: Always nil (returns nil result if not found).
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
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  cred (*configv1.Credential): The credential object to save.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Updates the internal credentials map.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential deletes a credential by ID.
//
// Parameters:
//  _ (context.Context): Unused context (required by interface).
//  id (string): The unique identifier of the credential to delete.
//
// Returns:
//  error: Always nil.
//
// Side Effects:
//  Removes the credential from the internal map.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
