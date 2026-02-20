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

// NewStore creates a new memory store.
//
// Summary: Initializes a new in-memory storage backend.
//
// Returns:
//   *Store: A pointer to the initialized Store instance.
//
// Throws/Errors:
//   None.
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
// Summary: Persists a user token for a specific service.
//
// Parameters:
//   - ctx: context.Context. The context for the operation (unused in memory store).
//   - token: *configv1.UserToken. The token object containing user ID, service ID, and token data.
//
// Returns:
//   error: Always returns nil for memory store.
//
// Throws/Errors:
//   None.
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
// Summary: Retrieves a stored token for a given user and service.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   *configv1.UserToken: The retrieved token, or nil if not found.
//   error: Always returns nil for memory store.
//
// Throws/Errors:
//   None.
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

func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

func (s *Store) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		list = append(list, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}
	return list, nil
}

func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) HasConfigSources() bool {
	return true
}

func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

func (s *Store) ListSecrets(_ context.Context) ([]*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Secret, 0, len(s.secrets))
	for _, secret := range s.secrets {
		list = append(list, proto.Clone(secret).(*configv1.Secret))
	}
	return list, nil
}

func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

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

func (s *Store) GetUser(_ context.Context, id string) (*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[id]; ok {
		return proto.Clone(user).(*configv1.User), nil
	}
	return nil, nil
}

func (s *Store) ListUsers(_ context.Context) ([]*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.User, 0, len(s.users))
	for _, user := range s.users {
		list = append(list, proto.Clone(user).(*configv1.User))
	}
	return list, nil
}

func (s *Store) UpdateUser(_ context.Context, user *configv1.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.GetId()]; !ok {
		return fmt.Errorf("user not found")
	}
	s.users[user.GetId()] = proto.Clone(user).(*configv1.User)
	return nil
}

func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

func (s *Store) ListProfiles(_ context.Context) ([]*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.ProfileDefinition, 0, len(s.profileDefinitions))
	for _, p := range s.profileDefinitions {
		list = append(list, proto.Clone(p).(*configv1.ProfileDefinition))
	}
	return list, nil
}

func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

func (s *Store) ListServiceCollections(_ context.Context) ([]*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Collection, 0, len(s.serviceCollections))
	for _, c := range s.serviceCollections {
		list = append(list, proto.Clone(c).(*configv1.Collection))
	}
	return list, nil
}

func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.Collection), nil
	}
	return nil, nil
}

func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

func (s *Store) ListCredentials(_ context.Context) ([]*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Credential, 0, len(s.credentials))
	for _, c := range s.credentials {
		list = append(list, proto.Clone(c).(*configv1.Credential))
	}
	return list, nil
}

func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
