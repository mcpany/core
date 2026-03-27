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

// NewStore provides newstore functionality.
//
// Summary: NewStore.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// SaveLog provides savelog functionality.
//
// Summary: SaveLog.
//
// Parameters.
//   - _: The parameter.
//   - entry: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveLog(_ context.Context, entry *logging.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, entry)
	return nil
}

// GetRecentLogs provides getrecentlogs functionality.
//
// Summary: GetRecentLogs.
//
// Parameters.
//   - _: The parameter.
//   - limit: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// SaveToken provides savetoken functionality.
//
// Summary: SaveToken.
//
// Parameters.
//   - _: The parameter.
//   - token: The parameter.
//
// Returns.
//   - result: The result.
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

// GetToken provides gettoken functionality.
//
// Summary: GetToken.
//
// Parameters.
//   - _: The parameter.
//   - userID: The parameter.
//   - serviceID: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// DeleteToken provides deletetoken functionality.
//
// Summary: DeleteToken.
//
// Parameters.
//   - _: The parameter.
//   - userID: The parameter.
//   - serviceID: The parameter.
//
// Returns.
//   - result: The result.
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

// Load provides load functionality.
//
// Summary: Load.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// SaveService provides saveservice functionality.
//
// Summary: SaveService.
//
// Parameters.
//   - _: The parameter.
//   - service: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// GetService provides getservice functionality.
//
// Summary: GetService.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

// ListServices provides listservices functionality.
//
// Summary: ListServices.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.UpstreamServiceConfig, 0, len(s.services))
	for _, svc := range s.services {
		list = append(list, proto.Clone(svc).(*configv1.UpstreamServiceConfig))
	}
	return list, nil
}

// DeleteService provides deleteservice functionality.
//
// Summary: DeleteService.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Store) Close() error {
	return nil
}

// HasConfigSources provides hasconfigsources functionality.
//
// Summary: HasConfigSources.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (s *Store) HasConfigSources() bool {
	return true
}

// GetGlobalSettings provides getglobalsettings functionality.
//
// Summary: GetGlobalSettings.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// SaveGlobalSettings provides saveglobalsettings functionality.
//
// Summary: SaveGlobalSettings.
//
// Parameters.
//   - _: The parameter.
//   - settings: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// ListSecrets provides listsecrets functionality.
//
// Summary: ListSecrets.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListSecrets(_ context.Context) ([]*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Secret, 0, len(s.secrets))
	for _, secret := range s.secrets {
		list = append(list, proto.Clone(secret).(*configv1.Secret))
	}
	return list, nil
}

// GetSecret provides getsecret functionality.
//
// Summary: GetSecret.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// SaveSecret provides savesecret functionality.
//
// Summary: SaveSecret.
//
// Parameters.
//   - _: The parameter.
//   - secret: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// DeleteSecret provides deletesecret functionality.
//
// Summary: DeleteSecret.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteSecret(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, id)
	return nil
}

// CreateUser provides createuser functionality.
//
// Summary: CreateUser.
//
// Parameters.
//   - _: The parameter.
//   - user: The parameter.
//
// Returns.
//   - result: The result.
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

// GetUser provides getuser functionality.
//
// Summary: GetUser.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetUser(_ context.Context, id string) (*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[id]; ok {
		return proto.Clone(user).(*configv1.User), nil
	}
	return nil, nil
}

// ListUsers provides listusers functionality.
//
// Summary: ListUsers.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListUsers(_ context.Context) ([]*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.User, 0, len(s.users))
	for _, user := range s.users {
		list = append(list, proto.Clone(user).(*configv1.User))
	}
	return list, nil
}

// UpdateUser provides updateuser functionality.
//
// Summary: UpdateUser.
//
// Parameters.
//   - _: The parameter.
//   - user: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) UpdateUser(_ context.Context, user *configv1.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.GetId()]; !ok {
		return fmt.Errorf("user not found")
	}
	s.users[user.GetId()] = proto.Clone(user).(*configv1.User)
	return nil
}

// DeleteUser provides deleteuser functionality.
//
// Summary: DeleteUser.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// ListProfiles provides listprofiles functionality.
//
// Summary: ListProfiles.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListProfiles(_ context.Context) ([]*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.ProfileDefinition, 0, len(s.profileDefinitions))
	for _, p := range s.profileDefinitions {
		list = append(list, proto.Clone(p).(*configv1.ProfileDefinition))
	}
	return list, nil
}

// GetProfile provides getprofile functionality.
//
// Summary: GetProfile.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

// SaveProfile provides saveprofile functionality.
//
// Summary: SaveProfile.
//
// Parameters.
//   - _: The parameter.
//   - profile: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// DeleteProfile provides deleteprofile functionality.
//
// Summary: DeleteProfile.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// ListServiceCollections provides listservicecollections functionality.
//
// Summary: ListServiceCollections.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListServiceCollections(_ context.Context) ([]*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Collection, 0, len(s.serviceCollections))
	for _, c := range s.serviceCollections {
		list = append(list, proto.Clone(c).(*configv1.Collection))
	}
	return list, nil
}

// GetServiceCollection provides getservicecollection functionality.
//
// Summary: GetServiceCollection.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.Collection), nil
	}
	return nil, nil
}

// SaveServiceCollection provides saveservicecollection functionality.
//
// Summary: SaveServiceCollection.
//
// Parameters.
//   - _: The parameter.
//   - collection: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// DeleteServiceCollection provides deleteservicecollection functionality.
//
// Summary: DeleteServiceCollection.
//
// Parameters.
//   - _: The parameter.
//   - name: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// ListCredentials provides listcredentials functionality.
//
// Summary: ListCredentials.
//
// Parameters.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) ListCredentials(_ context.Context) ([]*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*configv1.Credential, 0, len(s.credentials))
	for _, c := range s.credentials {
		list = append(list, proto.Clone(c).(*configv1.Credential))
	}
	return list, nil
}

// GetCredential provides getcredential functionality.
//
// Summary: GetCredential.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

// SaveCredential provides savecredential functionality.
//
// Summary: SaveCredential.
//
// Parameters.
//   - _: The parameter.
//   - cred: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// DeleteCredential provides deletecredential functionality.
//
// Summary: DeleteCredential.
//
// Parameters.
//   - _: The parameter.
//   - id: The parameter.
//
// Returns.
//   - result: The result.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
