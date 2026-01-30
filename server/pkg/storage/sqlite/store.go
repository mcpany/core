// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Store implements config.Store using SQLite.
type Store struct {
	db    *DB
	cache sync.Map
}

// NewStore creates a new SQLite store.
//
// db is the db.
//
// Returns the result.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// Close closes the underlying database connection.
//
// Returns an error if the operation fails.
func (s *Store) Close() error {
	return s.db.Close()
}

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
// For DB stores, we assume they always have a source (the DB itself).
func (s *Store) HasConfigSources() bool {
	return true
}

// Load implements config.Store interface.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	// 1. Load services
	services, err := s.ListServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// 2. Load users
	users, err := s.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// 3. Load Global Settings
	settings, err := s.GetGlobalSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global settings: %w", err)
	}

	// 4. Load Collections
	collections, err := s.ListServiceCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list service collections: %w", err)
	}

	builder := configv1.McpAnyServerConfig_builder{
		UpstreamServices: services,
		Users:            users,
		Collections:      collections,
	}
	if settings != nil {
		builder.GlobalSettings = settings
	}

	return builder.Build(), nil
}

// SaveService saves an upstream service configuration.
//
// ctx is the context for the request.
// service is the service instance.
//
// Returns an error if the operation fails.
func (s *Store) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if service.GetName() == "" {
		return fmt.Errorf("service name is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service config: %w", err)
	}

	query := `
	INSERT INTO upstream_services (id, name, config_json, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(name) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	id := service.GetId()
	if id == "" {
		id = service.GetName() // fallback
	}

	_, err = s.db.ExecContext(ctx, query, id, service.GetName(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save service: %w", err)
	}
	s.cache.Delete("services")
	return nil
}

// GetService retrieves an upstream service configuration by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	query := "SELECT config_json FROM upstream_services WHERE name = ?"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var service configv1.UpstreamServiceConfig
	if err := protojson.Unmarshal([]byte(configJSON), &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service config: %w", err)
	}
	return &service, nil
}

// ListServices lists all upstream service configurations.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	// ⚡ BOLT: Check cache first to avoid expensive deserialization
	// Randomized Selection from Top 5 High-Impact Targets
	if v, ok := s.cache.Load("services"); ok {
		return v.([]*configv1.UpstreamServiceConfig), nil
	}

	query := "SELECT config_json FROM upstream_services"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var services []*configv1.UpstreamServiceConfig
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var service configv1.UpstreamServiceConfig
		if err := opts.Unmarshal(configJSON, &service); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service config: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("services", services)
	return services, nil
}

// DeleteService deletes an upstream service configuration by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	s.cache.Delete("services")
	return nil
}

// GetGlobalSettings retrieves the global configuration.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	if v, ok := s.cache.Load("global_settings"); ok {
		return v.(*configv1.GlobalSettings), nil
	}

	query := "SELECT config_json FROM global_settings WHERE id = 1"
	row := s.db.QueryRowContext(ctx, query)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan global settings: %w", err)
	}

	var settings configv1.GlobalSettings
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := opts.Unmarshal([]byte(configJSON), &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global settings: %w", err)
	}

	s.cache.Store("global_settings", &settings)
	return &settings, nil
}

// SaveGlobalSettings saves the global configuration.
//
// ctx is the context for the request.
// settings is the settings.
//
// Returns an error if the operation fails.
func (s *Store) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal global settings: %w", err)
	}

	query := `
	INSERT INTO global_settings (id, config_json, updated_at)
	VALUES (1, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save global settings: %w", err)
	}
	s.cache.Delete("global_settings")
	return nil
}

// Users

// CreateUser creates a new user.
//
// ctx is the context for the request.
// user is the user.
//
// Returns an error if the operation fails.
func (s *Store) CreateUser(ctx context.Context, user *configv1.User) error {
	if user.GetId() == "" {
		return fmt.Errorf("user ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user config: %w", err)
	}

	query := `
	INSERT INTO users (id, config_json, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`
	_, err = s.db.ExecContext(ctx, query, user.GetId(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	s.cache.Delete("users")
	return nil
}

// GetUser retrieves a user by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	query := "SELECT config_json FROM users WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan user config: %w", err)
	}

	var user configv1.User
	if err := protojson.Unmarshal([]byte(configJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
	}
	return &user, nil
}

// ListUsers retrieves all users.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	if v, ok := s.cache.Load("users"); ok {
		return v.([]*configv1.User), nil
	}

	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []*configv1.User
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan user config: %w", err)
		}

		var user configv1.User
		if err := opts.Unmarshal([]byte(configJSON), &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("users", users)
	return users, nil
}

// UpdateUser updates an existing user.
//
// ctx is the context for the request.
// user is the user.
//
// Returns an error if the operation fails.
func (s *Store) UpdateUser(ctx context.Context, user *configv1.User) error {
	if user.GetId() == "" {
		return fmt.Errorf("user ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user config: %w", err)
	}

	query := `
	UPDATE users
	SET config_json = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`
	res, err := s.db.ExecContext(ctx, query, string(configJSON), user.GetId())
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	s.cache.Delete("users")
	return nil
}

// DeleteUser deletes a user by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns an error if the operation fails.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	s.cache.Delete("users")
	return nil
}

// Secrets

// ListSecrets retrieves all secrets.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	if v, ok := s.cache.Load("secrets"); ok {
		return v.([]*configv1.Secret), nil
	}

	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM secrets")
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var secrets []*configv1.Secret
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var secret configv1.Secret
		if err := opts.Unmarshal([]byte(configJSON), &secret); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
		}
		secrets = append(secrets, &secret)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("secrets", secrets)
	return secrets, nil
}

// GetSecret retrieves a secret by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	query := "SELECT config_json FROM secrets WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan secret: %w", err)
	}

	var secret configv1.Secret
	if err := protojson.Unmarshal([]byte(configJSON), &secret); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}
	return &secret, nil
}

// SaveSecret saves a secret.
//
// ctx is the context for the request.
// secret is the secret.
//
// Returns an error if the operation fails.
func (s *Store) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	if secret.GetId() == "" {
		return fmt.Errorf("secret id is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(secret)
	if err != nil {
		return fmt.Errorf("failed to marshal secret: %w", err)
	}

	query := `
	INSERT INTO secrets (id, name, key, config_json, updated_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		key = excluded.key,
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, secret.GetId(), secret.GetName(), secret.GetKey(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	s.cache.Delete("secrets")
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns an error if the operation fails.
func (s *Store) DeleteSecret(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM secrets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	s.cache.Delete("secrets")
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	if v, ok := s.cache.Load("profiles"); ok {
		return v.([]*configv1.ProfileDefinition), nil
	}

	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM profile_definitions")
	if err != nil {
		return nil, fmt.Errorf("failed to query profile_definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*configv1.ProfileDefinition
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var profile configv1.ProfileDefinition
		if err := opts.Unmarshal([]byte(configJSON), &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile config: %w", err)
		}
		profiles = append(profiles, &profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("profiles", profiles)
	return profiles, nil
}

// GetProfile retrieves a profile definition by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	query := "SELECT config_json FROM profile_definitions WHERE name = ?"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var profile configv1.ProfileDefinition
	if err := protojson.Unmarshal([]byte(configJSON), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile config: %w", err)
	}
	return &profile, nil
}

// SaveProfile saves a profile definition.
//
// ctx is the context for the request.
// profile is the profile.
//
// Returns an error if the operation fails.
func (s *Store) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	if profile.GetName() == "" {
		return fmt.Errorf("profile name is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile config: %w", err)
	}

	query := `
	INSERT INTO profile_definitions (id, name, config_json, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(name) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	id := profile.GetName()

	_, err = s.db.ExecContext(ctx, query, id, profile.GetName(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}
	s.cache.Delete("profiles")
	return nil
}

// DeleteProfile deletes a profile definition by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM profile_definitions WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	s.cache.Delete("profiles")
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	if v, ok := s.cache.Load("service_collections"); ok {
		return v.([]*configv1.Collection), nil
	}

	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_collections")
	if err != nil {
		return nil, fmt.Errorf("failed to query service_collections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var collections []*configv1.Collection
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var collection configv1.Collection
		if err := opts.Unmarshal([]byte(configJSON), &collection); err != nil {
			return nil, fmt.Errorf("failed to unmarshal collection config: %w", err)
		}
		collections = append(collections, &collection)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("service_collections", collections)
	return collections, nil
}

// GetServiceCollection retrieves a service collection by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	query := "SELECT config_json FROM service_collections WHERE name = ?"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var collection configv1.Collection
	if err := protojson.Unmarshal([]byte(configJSON), &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collection config: %w", err)
	}
	return &collection, nil
}

// SaveServiceCollection saves a service collection.
//
// ctx is the context for the request.
// collection is the collection.
//
// Returns an error if the operation fails.
func (s *Store) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error {
	if collection.GetName() == "" {
		return fmt.Errorf("collection name is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(collection)
	if err != nil {
		return fmt.Errorf("failed to marshal collection config: %w", err)
	}

	query := `
	INSERT INTO service_collections (id, name, config_json, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(name) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	id := collection.GetName()

	_, err = s.db.ExecContext(ctx, query, id, collection.GetName(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save collection: %w", err)
	}
	s.cache.Delete("service_collections")
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteServiceCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_collections WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	s.cache.Delete("service_collections")
	return nil
}

// Tokens

// SaveToken saves a user token.
//
// ctx is the context for the request.
// token is the token.
//
// Returns an error if the operation fails.
func (s *Store) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	if token.GetUserId() == "" || token.GetServiceId() == "" {
		return fmt.Errorf("user ID and service ID are required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	query := `
	INSERT INTO user_tokens (user_id, service_id, config_json, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(user_id, service_id) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, token.GetUserId(), token.GetServiceId(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	return nil
}

// GetToken retrieves a user token by user ID and service ID.
//
// ctx is the context for the request.
// userID is the userID.
// serviceID is the serviceID.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	query := "SELECT config_json FROM user_tokens WHERE user_id = ? AND service_id = ?"
	row := s.db.QueryRowContext(ctx, query, userID, serviceID)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan token config: %w", err)
	}

	var token configv1.UserToken
	if err := protojson.Unmarshal([]byte(configJSON), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return &token, nil
}

// DeleteToken deletes a user token.
//
// ctx is the context for the request.
// userID is the userID.
// serviceID is the serviceID.
//
// Returns an error if the operation fails.
func (s *Store) DeleteToken(ctx context.Context, userID, serviceID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = ? AND service_id = ?", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// Credentials

// ListCredentials retrieves all credentials.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	if v, ok := s.cache.Load("credentials"); ok {
		return v.([]*configv1.Credential), nil
	}

	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM credentials")
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var credentials []*configv1.Credential
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan credential config: %w", err)
		}

		var cred configv1.Credential
		if err := opts.Unmarshal([]byte(configJSON), &cred); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
		}
		credentials = append(credentials, &cred)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	s.cache.Store("credentials", credentials)
	return credentials, nil
}

// GetCredential retrieves a credential by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	query := "SELECT config_json FROM credentials WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	var cred configv1.Credential
	if err := protojson.Unmarshal([]byte(configJSON), &cred); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return &cred, nil
}

// SaveCredential saves a credential.
//
// ctx is the context for the request.
// cred is the cred.
//
// Returns an error if the operation fails.
func (s *Store) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	if cred.GetId() == "" {
		return fmt.Errorf("credential ID is required")
	}

	opts := protojson.MarshalOptions{UseProtoNames: true}
	configJSON, err := opts.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal credential: %w", err)
	}

	query := `
	INSERT INTO credentials (id, name, config_json, updated_at)
	VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, cred.GetId(), cred.GetName(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save credential: %w", err)
	}
	s.cache.Delete("credentials")
	return nil
}

// DeleteCredential deletes a credential by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns an error if the operation fails.
func (s *Store) DeleteCredential(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	s.cache.Delete("credentials")
	return nil
}
