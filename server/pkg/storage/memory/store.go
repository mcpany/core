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

// Summary: Store implements storage.Storage in memory. A thread-safe, in-memory implementation of the Storage interface, primarily for testing.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: NewStore creates a new memory store. Initializes a new, empty in-memory store.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Store: The resulting *Store.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: SaveLog saves a log entry. Appends a log entry to the in-memory log store.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - entry (*logging.LogEntry): The entry parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveLog(_ context.Context, entry *logging.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, entry)
	return nil
}

// Summary: GetRecentLogs retrieves recent log entries. Returns the N most recent log entries.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - limit (int): The limit parameter.
//
// Returns:
//   - []*logging.LogEntry: The resulting []*logging.LogEntry.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: SaveToken saves a user token. Stores a user token in memory.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - token (*configv1.UserToken): The token parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetToken retrieves a user token by user ID and service ID. Retrieves a stored user token.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - userID (string): The userID parameter.
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - *configv1.UserToken: The resulting *configv1.UserToken.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: DeleteToken deletes a user token. Removes a user token from memory.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - userID (string): The userID parameter.
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: Load retrieves the full server configuration. Constructs and returns the complete server configuration from stored components.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The resulting *configv1.McpAnyServerConfig.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: SaveService saves a single upstream service configuration. Stores an upstream service configuration.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - service (*configv1.UpstreamServiceConfig): The service parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveService(_ context.Context, service *configv1.UpstreamServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service.GetName()] = proto.Clone(service).(*configv1.UpstreamServiceConfig)
	return nil
}

// Summary: GetService retrieves a single upstream service configuration by name. Retrieves an upstream service configuration.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The resulting *configv1.UpstreamServiceConfig.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetService(_ context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if svc, ok := s.services[name]; ok {
		return proto.Clone(svc).(*configv1.UpstreamServiceConfig), nil
	}
	return nil, nil
}

// Summary: ListServices lists all upstream service configurations. Lists all stored upstream service configurations.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The resulting []*configv1.UpstreamServiceConfig.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: DeleteService deletes an upstream service configuration by name. Deletes an upstream service configuration.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) DeleteService(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
	return nil
}

// Summary: Close closes the underlying storage connection. No-op for in-memory store.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) Close() error {
	return nil
}

// Summary: HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured. Indicates if the store supports config sources (always true for this mock).
//
// Parameters:
//   - None.
//
// Returns:
//   - bool: The resulting bool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (s *Store) HasConfigSources() bool {
	return true
}

// Summary: GetGlobalSettings retrieves the global configuration. Retrieves the global settings object.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - *configv1.GlobalSettings: The resulting *configv1.GlobalSettings.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetGlobalSettings(_ context.Context) (*configv1.GlobalSettings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.globalSettings == nil {
		return &configv1.GlobalSettings{}, nil
	}
	return proto.Clone(s.globalSettings).(*configv1.GlobalSettings), nil
}

// Summary: SaveGlobalSettings saves the global configuration. Persists the global settings.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - settings (*configv1.GlobalSettings): The settings parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveGlobalSettings(_ context.Context, settings *configv1.GlobalSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = proto.Clone(settings).(*configv1.GlobalSettings)
	return nil
}

// Summary: ListSecrets retrieves all secrets. Lists all stored secrets.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.Secret: The resulting []*configv1.Secret.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: GetSecret retrieves a secret by ID. Retrieves a secret by its ID.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - *configv1.Secret: The resulting *configv1.Secret.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetSecret(_ context.Context, id string) (*configv1.Secret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if secret, ok := s.secrets[id]; ok {
		return proto.Clone(secret).(*configv1.Secret), nil
	}
	return nil, nil
}

// Summary: SaveSecret saves a secret. Stores a secret.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - secret (*configv1.Secret): The secret parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveSecret(_ context.Context, secret *configv1.Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[secret.GetId()] = proto.Clone(secret).(*configv1.Secret)
	return nil
}

// Summary: DeleteSecret deletes a secret by ID. Deletes a secret.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetUser retrieves a user by ID. Retrieves a user.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - *configv1.User: The resulting *configv1.User.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetUser(_ context.Context, id string) (*configv1.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[id]; ok {
		return proto.Clone(user).(*configv1.User), nil
	}
	return nil, nil
}

// Summary: ListUsers retrieves all users. Lists all users.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.User: The resulting []*configv1.User.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: DeleteUser deletes a user by ID. Deletes a user.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.users, id)
	return nil
}

// Profiles

// Summary: ListProfiles retrieves all profile definitions. Lists all stored profile definitions.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.ProfileDefinition: The resulting []*configv1.ProfileDefinition.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: GetProfile retrieves a profile definition by name. Retrieves a profile by name.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - *configv1.ProfileDefinition: The resulting *configv1.ProfileDefinition.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetProfile(_ context.Context, name string) (*configv1.ProfileDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.profileDefinitions[name]; ok {
		return proto.Clone(p).(*configv1.ProfileDefinition), nil
	}
	return nil, nil
}

// Summary: SaveProfile saves a profile definition. Stores a profile definition.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - profile (*configv1.ProfileDefinition): The profile parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveProfile(_ context.Context, profile *configv1.ProfileDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.profileDefinitions[profile.GetName()] = proto.Clone(profile).(*configv1.ProfileDefinition)
	return nil
}

// Summary: DeleteProfile deletes a profile definition by name. Deletes a profile.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) DeleteProfile(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.profileDefinitions, name)
	return nil
}

// Service Collections

// Summary: ListServiceCollections retrieves all service collections. Lists all service collections.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.Collection: The resulting []*configv1.Collection.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: GetServiceCollection retrieves a service collection by name. Retrieves a service collection.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - *configv1.Collection: The resulting *configv1.Collection.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetServiceCollection(_ context.Context, name string) (*configv1.Collection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.serviceCollections[name]; ok {
		return proto.Clone(c).(*configv1.Collection), nil
	}
	return nil, nil
}

// Summary: SaveServiceCollection saves a service collection. Stores a service collection.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - collection (*configv1.Collection): The collection parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveServiceCollection(_ context.Context, collection *configv1.Collection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serviceCollections[collection.GetName()] = proto.Clone(collection).(*configv1.Collection)
	return nil
}

// Summary: DeleteServiceCollection deletes a service collection by name. Deletes a service collection.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - name (string): The name parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) DeleteServiceCollection(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.serviceCollections, name)
	return nil
}

// Credentials

// Summary: ListCredentials retrieves all credentials. Lists all credentials.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - []*configv1.Credential: The resulting []*configv1.Credential.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
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

// Summary: GetCredential retrieves a credential by ID. Retrieves a credential by ID.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - *configv1.Credential: The resulting *configv1.Credential.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) GetCredential(_ context.Context, id string) (*configv1.Credential, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.credentials[id]; ok {
		return proto.Clone(c).(*configv1.Credential), nil
	}
	return nil, nil
}

// Summary: SaveCredential saves a credential. Stores a credential.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - cred (*configv1.Credential): The cred parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) SaveCredential(_ context.Context, cred *configv1.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[cred.GetId()] = proto.Clone(cred).(*configv1.Credential)
	return nil
}

// Summary: DeleteCredential deletes a credential by ID. Deletes a credential.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - id (string): The id parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (s *Store) DeleteCredential(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.credentials, id)
	return nil
}
