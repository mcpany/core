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

// Store implements storage.Storage in memory.
type Store struct {
	mu                 sync.RWMutex
	services           map[string]*configv1.UpstreamServiceConfig
	secrets            map[string]*configv1.Secret
	users              map[string]*configv1.User
	profileDefinitions map[string]*configv1.ProfileDefinition
	serviceCollections map[string]*configv1.UpstreamServiceCollectionShare
	globalSettings     *configv1.GlobalSettings
	tokens             map[string]*configv1.UserToken
	credentials        map[string]*configv1.Credential
}

// NewStore creates a new memory store.
func NewStore() *Store {
	return &Store{
		services:           make(map[string]*configv1.UpstreamServiceConfig),
		secrets:            make(map[string]*configv1.Secret),
		users:              make(map[string]*configv1.User),
		profileDefinitions: make(map[string]*configv1.ProfileDefinition),
		serviceCollections: make(map[string]*configv1.UpstreamServiceCollectionShare),
		tokens:             make(map[string]*configv1.UserToken),
		credentials:        make(map[string]*configv1.Credential),
	}
}

// Helper to generate token key.
func tokenKey(userID, serviceID string) string {
	return userID + "|" + serviceID
}

// SaveToken saves a user token.
func (s *Store) SaveToken(_ context.Context, token *configv1.UserToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := tokenKey(token.GetUserId(), token.GetServiceId())
	s.tokens[key] = proto.Clone(token).(*configv1.UserToken)
	return nil
}

// GetToken retrieves a user token by user ID and service ID.
func (s *Store) GetToken(_ context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := tokenKey(userID, serviceID)
	if token, ok := s.tokens[key]; ok {
		return proto.Clone(token).(*configv1.UserToken), nil
	}
	return nil, nil
}

// DeleteToken deletes a user token.
func (s *Store) DeleteToken(_ context.Context, userID, serviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := tokenKey(userID, serviceID)
	delete(s.tokens, key)
	return nil
}

// Load retrieves the full server configuration.
func (s *Store) Load(_ context.Context) (*configv1.McpAnyServerConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg := &configv1.McpAnyServerConfig{}
	// Pre-allocate slice if we knew the size, but here we iterate map.
	// We can count map keys.
	cfg.UpstreamServices = make([]*configv1.UpstreamServiceConfig, 0, len(s.services))

	for _, svc := range s.services {
		cfg.UpstreamServices = append(cfg.UpstreamServices, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}

	if len(s.profileDefinitions) > 0 {
		cfg.GlobalSettings = &configv1.GlobalSettings{} // Ensure GlobalSettings structure exists if we have profiles?
		// But s.globalSettings might override it.
		// Wait, profiles are inside GlobalSettings.
		// So if s.globalSettings is nil, we should create it if we have profiles?
		// Or just assume s.globalSettings tracks everything?
		// In memory store, we separated `profileDefinitions` map.
		// So we should merge them into `cfg`.
	}

	if s.globalSettings != nil {
		cfg.GlobalSettings = proto.Clone(s.globalSettings).(*configv1.GlobalSettings)
	} else if len(s.profileDefinitions) > 0 {
		cfg.GlobalSettings = &configv1.GlobalSettings{}
	}

	if cfg.GlobalSettings != nil {
		// Populate profiles from separate map if not already in GlobalSettings?
		// Memory store usually keeps source of truth in maps.
		// But SaveGlobalSettings saves the whole blob.
		// If we use separate map for profiles, we should inject them.
		// But SaveGlobalSettings might overwrite profiles if it contains them?
		// The design of MemoryStore:
		// Logic: `SaveProfile` updates `profileDefinitions`. `SaveGlobalSettings` updates `globalSettings`.
		// If `GlobalSettings` contains `ProfileDefinitions`, they are duplicated?
		// Ideally `Load` should combine them.
		// Let's append `profileDefinitions` map to `cfg.GlobalSettings.ProfileDefinitions`.
		for _, p := range s.profileDefinitions {
			cfg.GlobalSettings.ProfileDefinitions = append(cfg.GlobalSettings.ProfileDefinitions, proto.Clone(p).(*configv1.ProfileDefinition))
		}
	}

	return cfg, nil
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

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
// Memory store always has data available.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings retrieves the global configuration.
func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// SaveGlobalSettings saves the global configuration.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets retrieves all secrets.
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
func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// SaveSecret saves a secret.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret deletes a secret by ID.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser creates a new user.
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
func (s *Store) GetUser(_ context.Context, id string) (*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[id]; ok {
		return proto.Clone(user).(*configv1.User), nil
	}
	return nil, nil
}

// ListUsers retrieves all users.
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
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
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
func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

// SaveProfile saves a profile definition.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile deletes a profile definition by name.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
func (s *Store) ListServiceCollections(_ context.Context) ([]*configv1.UpstreamServiceCollectionShare, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.UpstreamServiceCollectionShare, 0, len(s.serviceCollections))
	for _, c := range s.serviceCollections {
		list = append(list, proto.Clone(c).(*configv1.UpstreamServiceCollectionShare))
	}
	return list, nil
}

// GetServiceCollection retrieves a service collection by name.
func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.UpstreamServiceCollectionShare, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.UpstreamServiceCollectionShare), nil
	}
	return nil, nil
}

// SaveServiceCollection saves a service collection.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.UpstreamServiceCollectionShare) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.UpstreamServiceCollectionShare)
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// ListCredentials retrieves all credentials.
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
func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

// SaveCredential saves a credential.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential deletes a credential by ID.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
