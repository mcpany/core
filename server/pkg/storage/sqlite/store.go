// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Store implements config.Store using SQLite.
type Store struct {
	db *DB
}

// NewStore creates a new SQLite store.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Load implements config.Store interface.
func (s *Store) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM upstream_services")
	if err != nil {
		return nil, fmt.Errorf("failed to query upstream_services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var services []*configv1.UpstreamServiceConfig
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var service configv1.UpstreamServiceConfig
		// Allow unknown fields to be safe against schema evolution
		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		if err := opts.Unmarshal([]byte(configJSON), &service); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service config: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &configv1.McpAnyServerConfig{
		UpstreamServices: services,
	}, nil
}

// SaveService saves an upstream service configuration.
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
	return nil
}

// GetService retrieves an upstream service configuration by name.
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
func (s *Store) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	cfg, err := s.Load(ctx)
	if err != nil {
		return nil, err
	}
	return cfg.UpstreamServices, nil
}

// DeleteService deletes an upstream service configuration by name.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetGlobalSettings retrieves the global configuration.
func (s *Store) GetGlobalSettings() (*configv1.GlobalSettings, error) {
	query := "SELECT config_json FROM global_settings WHERE id = 1"
	row := s.db.QueryRowContext(context.TODO(), query)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan global settings: %w", err)
	}

	var settings configv1.GlobalSettings
	if err := protojson.Unmarshal([]byte(configJSON), &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global settings: %w", err)
	}
	return &settings, nil
}

// SaveGlobalSettings saves the global configuration.
func (s *Store) SaveGlobalSettings(settings *configv1.GlobalSettings) error {
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
	_, err = s.db.ExecContext(context.TODO(), query, string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save global settings: %w", err)
	}
	return nil
}

// Users

// CreateUser creates a new user.
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
	return nil
}

// GetUser retrieves a user by ID.
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
func (s *Store) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []*configv1.User
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan user config: %w", err)
		}

		var user configv1.User
		if err := protojson.Unmarshal([]byte(configJSON), &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// UpdateUser updates an existing user.
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
	return nil
}

// DeleteUser deletes a user by ID.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Secrets

// ListSecrets retrieves all secrets.
func (s *Store) ListSecrets() ([]*configv1.Secret, error) {
	rows, err := s.db.QueryContext(context.TODO(), "SELECT config_json FROM secrets")
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var secrets []*configv1.Secret
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var secret configv1.Secret
		if err := protojson.Unmarshal([]byte(configJSON), &secret); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
		}
		secrets = append(secrets, &secret)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return secrets, nil
}

// GetSecret retrieves a secret by ID.
func (s *Store) GetSecret(id string) (*configv1.Secret, error) {
	query := "SELECT config_json FROM secrets WHERE id = ?"
	row := s.db.QueryRowContext(context.TODO(), query, id)

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
func (s *Store) SaveSecret(secret *configv1.Secret) error {
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
	_, err = s.db.ExecContext(context.TODO(), query, secret.GetId(), secret.GetName(), secret.GetKey(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}
	return nil
}

// DeleteSecret deletes a secret by ID.
func (s *Store) DeleteSecret(id string) error {
	_, err := s.db.ExecContext(context.TODO(), "DELETE FROM secrets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
func (s *Store) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM profile_definitions")
	if err != nil {
		return nil, fmt.Errorf("failed to query profile_definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*configv1.ProfileDefinition
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var profile configv1.ProfileDefinition
		if err := protojson.Unmarshal([]byte(configJSON), &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal profile config: %w", err)
		}
		profiles = append(profiles, &profile)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return profiles, nil
}

// GetProfile retrieves a profile definition by name.
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
	return nil
}

// DeleteProfile deletes a profile definition by name.
func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM profile_definitions WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
func (s *Store) ListServiceCollections(ctx context.Context) ([]*configv1.UpstreamServiceCollectionShare, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_collections")
	if err != nil {
		return nil, fmt.Errorf("failed to query service_collections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var collections []*configv1.UpstreamServiceCollectionShare
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var collection configv1.UpstreamServiceCollectionShare
		if err := protojson.Unmarshal([]byte(configJSON), &collection); err != nil {
			return nil, fmt.Errorf("failed to unmarshal collection config: %w", err)
		}
		collections = append(collections, &collection)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return collections, nil
}

// GetServiceCollection retrieves a service collection by name.
func (s *Store) GetServiceCollection(ctx context.Context, name string) (*configv1.UpstreamServiceCollectionShare, error) {
	query := "SELECT config_json FROM service_collections WHERE name = ?"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON string
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var collection configv1.UpstreamServiceCollectionShare
	if err := protojson.Unmarshal([]byte(configJSON), &collection); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collection config: %w", err)
	}
	return &collection, nil
}

// SaveServiceCollection saves a service collection.
func (s *Store) SaveServiceCollection(ctx context.Context, collection *configv1.UpstreamServiceCollectionShare) error {
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
	return nil
}

// DeleteServiceCollection deletes a service collection by name.
func (s *Store) DeleteServiceCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_collections WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

// Tokens

// SaveToken saves a user token.
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
func (s *Store) DeleteToken(ctx context.Context, userID, serviceID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = ? AND service_id = ?", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
