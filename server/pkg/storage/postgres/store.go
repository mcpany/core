// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	loggingv1 "github.com/mcpany/core/proto/logging/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Store implements config.Store using PostgreSQL.
type Store struct {
	db *DB
}

// NewStore creates a new PostgreSQL store.
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
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM upstream_services")
	if err != nil {
		return nil, fmt.Errorf("failed to query upstream_services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var services []*configv1.UpstreamServiceConfig
	// Allow unknown fields to be safe against schema evolution
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
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	// 2. Load users
	userRows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = userRows.Close() }()

	var users []*configv1.User
	for userRows.Next() {
		var configJSON []byte
		if err := userRows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan user config_json: %w", err)
		}

		var user configv1.User
		if err := opts.Unmarshal(configJSON, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
		}
		users = append(users, &user)
	}
	_ = userRows.Close()
	if err := userRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user rows: %w", err)
	}

	// 3. Load Global Settings
	var settings *configv1.GlobalSettings
	settingsRow := s.db.QueryRowContext(ctx, "SELECT config_json FROM global_settings WHERE id = 1")
	var settingsJSON []byte
	if err := settingsRow.Scan(&settingsJSON); err == nil {
		var s configv1.GlobalSettings
		if err := opts.Unmarshal(settingsJSON, &s); err == nil {
			settings = &s
		}
	}

	// 4. Load Collections
	collectionRows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_collections")
	var collections []*configv1.Collection
	if err == nil {
		defer func() { _ = collectionRows.Close() }()
		for collectionRows.Next() {
			var configJSON []byte
			if err := collectionRows.Scan(&configJSON); err != nil {
				continue
			}
			var c configv1.Collection
			if err := opts.Unmarshal(configJSON, &c); err == nil {
				collections = append(collections, &c)
			}
		}
		if err := collectionRows.Err(); err != nil {
			return nil, fmt.Errorf("failed to iterate collection rows: %w", err)
		}
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
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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

// Logs

// SaveLog saves a log entry.
func (s *Store) SaveLog(ctx context.Context, log *loggingv1.LogEntry) error {
	query := `
	INSERT INTO logs (id, timestamp, level, source, message, metadata_json, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`
	_, err := s.db.ExecContext(ctx, query,
		log.GetId(),
		log.GetTimestamp(),
		log.GetLevel(),
		log.GetSource(),
		log.GetMessage(),
		log.GetMetadataJson(),
	)
	if err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

// ListLogs lists log entries.
func (s *Store) ListLogs(ctx context.Context, filter *loggingv1.LogFilter) ([]*loggingv1.LogEntry, error) {
	query := "SELECT id, timestamp, level, source, message, metadata_json FROM logs WHERE 1=1"
	var args []interface{}
	paramIdx := 1

	if filter != nil {
		if filter.GetSource() != "" && filter.GetSource() != "ALL" {
			query += fmt.Sprintf(" AND source = $%d", paramIdx)
			args = append(args, filter.GetSource())
			paramIdx++
		}
		if filter.GetLevel() != "" && filter.GetLevel() != "ALL" {
			query += fmt.Sprintf(" AND level = $%d", paramIdx)
			args = append(args, filter.GetLevel())
			paramIdx++
		}
		if filter.GetSearch() != "" {
			query += fmt.Sprintf(" AND (message LIKE $%d OR source LIKE $%d)", paramIdx, paramIdx+1)
			term := "%" + filter.GetSearch() + "%"
			args = append(args, term, term)
			paramIdx += 2
		}
		if filter.GetStartTime() != "" {
			query += fmt.Sprintf(" AND timestamp >= $%d", paramIdx)
			args = append(args, filter.GetStartTime())
			paramIdx++
		}
		if filter.GetEndTime() != "" {
			query += fmt.Sprintf(" AND timestamp <= $%d", paramIdx)
			args = append(args, filter.GetEndTime())
			paramIdx++
		}
	}

	query += " ORDER BY timestamp DESC"

	if filter != nil && filter.GetLimit() > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIdx)
		args = append(args, filter.GetLimit())
	} else {
		query += " LIMIT 1000" // Default limit
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []*loggingv1.LogEntry
	for rows.Next() {
		var l loggingv1.LogEntry
		// Ensure nullable fields are handled if necessary, but protobuf fields are not null in Go structs (strings are empty).
		// Postgres might return NULL for source/metadata_json if not set.
		var source, metadataJSON sql.NullString
		if err := rows.Scan(&l.Id, &l.Timestamp, &l.Level, &source, &l.Message, &metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		if source.Valid {
			l.Source = source.String
		}
		if metadataJSON.Valid {
			l.MetadataJson = metadataJSON.String
		}
		logs = append(logs, &l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}

	// Reverse logs to be chronological (oldest to newest)
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

// GetService retrieves an upstream service configuration by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	query := "SELECT config_json FROM upstream_services WHERE name = $1"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var service configv1.UpstreamServiceConfig
	if err := protojson.Unmarshal(configJSON, &service); err != nil {
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
	query := "SELECT config_json FROM upstream_services"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var services []*configv1.UpstreamServiceConfig
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var service configv1.UpstreamServiceConfig
		if err := protojson.Unmarshal(configJSON, &service); err != nil {
			return nil, fmt.Errorf("failed to unmarshal service config: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return services, nil
}

// DeleteService deletes an upstream service configuration by name.
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetGlobalSettings retrieves the global configuration.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	query := "SELECT config_json FROM global_settings WHERE id = 1"
	row := s.db.QueryRowContext(ctx, query)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan global settings: %w", err)
	}

	var settings configv1.GlobalSettings
	if err := protojson.Unmarshal(configJSON, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal global settings: %w", err)
	}
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
	VALUES (1, $1, CURRENT_TIMESTAMP)
	ON CONFLICT(id) DO UPDATE SET
		config_json = excluded.config_json,
		updated_at = excluded.updated_at;
	`
	_, err = s.db.ExecContext(ctx, query, string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to save global settings: %w", err)
	}
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
	VALUES ($1, $2, CURRENT_TIMESTAMP)
	`
	_, err = s.db.ExecContext(ctx, query, user.GetId(), string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
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
	query := "SELECT config_json FROM users WHERE id = $1"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan user config: %w", err)
	}

	var user configv1.User
	if err := protojson.Unmarshal(configJSON, &user); err != nil {
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
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []*configv1.User
	// Allow unknown fields to be safe against schema evolution
	opts := protojson.UnmarshalOptions{DiscardUnknown: true}

	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan user config: %w", err)
		}

		var user configv1.User
		if err := opts.Unmarshal(configJSON, &user); err != nil {
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
	SET config_json = $2, updated_at = CURRENT_TIMESTAMP
	WHERE id = $1
	`
	res, err := s.db.ExecContext(ctx, query, user.GetId(), string(configJSON))
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
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns an error if the operation fails.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
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
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM secrets")
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var secrets []*configv1.Secret
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var secret configv1.Secret
		if err := protojson.Unmarshal(configJSON, &secret); err != nil {
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
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	query := "SELECT config_json FROM secrets WHERE id = $1"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan secret: %w", err)
	}

	var secret configv1.Secret
	if err := protojson.Unmarshal(configJSON, &secret); err != nil {
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
	VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
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
	return nil
}

// DeleteSecret deletes a secret by ID.
//
// ctx is the context for the request.
// id is the unique identifier.
//
// Returns an error if the operation fails.
func (s *Store) DeleteSecret(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM secrets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
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
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM profile_definitions")
	if err != nil {
		return nil, fmt.Errorf("failed to query profile_definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*configv1.ProfileDefinition
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var profile configv1.ProfileDefinition
		if err := protojson.Unmarshal(configJSON, &profile); err != nil {
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
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	query := "SELECT config_json FROM profile_definitions WHERE name = $1"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var profile configv1.ProfileDefinition
	if err := protojson.Unmarshal(configJSON, &profile); err != nil {
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
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM profile_definitions WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
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
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_collections")
	if err != nil {
		return nil, fmt.Errorf("failed to query service_collections: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var collections []*configv1.Collection
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan config_json: %w", err)
		}

		var collection configv1.Collection
		if err := protojson.Unmarshal(configJSON, &collection); err != nil {
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
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns the result.
// Returns an error if the operation fails.
func (s *Store) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	query := "SELECT config_json FROM service_collections WHERE name = $1"
	row := s.db.QueryRowContext(ctx, query, name)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan config_json: %w", err)
	}

	var collection configv1.Collection
	if err := protojson.Unmarshal(configJSON, &collection); err != nil {
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
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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
//
// ctx is the context for the request.
// name is the name of the resource.
//
// Returns an error if the operation fails.
func (s *Store) DeleteServiceCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_collections WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
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
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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
	query := "SELECT config_json FROM user_tokens WHERE user_id = $1 AND service_id = $2"
	row := s.db.QueryRowContext(ctx, query, userID, serviceID)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan token config: %w", err)
	}

	var token configv1.UserToken
	if err := protojson.Unmarshal(configJSON, &token); err != nil {
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
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = $1 AND service_id = $2", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
