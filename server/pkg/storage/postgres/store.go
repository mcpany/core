// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Store implements config.Store using PostgreSQL.
//
// Summary: PostgreSQL implementation of the Storage interface.
type Store struct {
	db *DB
}

// NewStore creates a new PostgreSQL store.
//
// Summary: Initializes a new PostgreSQL storage instance.
//
// Parameters:
//   - db: *DB. The PostgreSQL database wrapper.
//
// Returns:
//   - *Store: A pointer to the newly created Store instance.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// Close closes the underlying database connection.
//
// Summary: Closes the PostgreSQL database connection.
//
// Returns:
//   - error: Returns nil on success, or an error if the connection close fails.
func (s *Store) Close() error {
	return s.db.Close()
}

// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
// For DB stores, we assume they always have a source (the DB itself).
//
// Summary: Checks if configuration sources are available.
//
// Returns:
//   - bool: True, as the database is considered a valid configuration source.
func (s *Store) HasConfigSources() bool {
	return true
}

// Load implements config.Store interface.
//
// Summary: Loads the complete server configuration from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The complete server configuration object.
//   - error: Returns nil on success, or an error if the operation fails.
func (s *Store) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	var (
		services    []*configv1.UpstreamServiceConfig
		users       []*configv1.User
		settings    *configv1.GlobalSettings
		collections []*configv1.Collection
		profiles    []*configv1.ProfileDefinition

		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	// ⚡ BOLT: Parallelized data loading (5 concurrent queries) to reduce latency.
	// Randomized Selection from Top 5 High-Impact Targets

	wg.Add(5)

	// 1. Load services
	go func() {
		defer wg.Done()
		rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM upstream_services")
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to query upstream_services: %w", err))
			mu.Unlock()
			return
		}
		defer func() { _ = rows.Close() }()

		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		for rows.Next() {
			var configJSON []byte
			if err := rows.Scan(&configJSON); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to scan config_json: %w", err))
				mu.Unlock()
				return
			}

			var service configv1.UpstreamServiceConfig
			if err := opts.Unmarshal(configJSON, &service); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to unmarshal service config: %w", err))
				mu.Unlock()
				return
			}
			services = append(services, &service)
		}
		if err := rows.Err(); err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to iterate rows: %w", err))
			mu.Unlock()
		}
	}()

	// 2. Load users
	go func() {
		defer wg.Done()
		userRows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to query users: %w", err))
			mu.Unlock()
			return
		}
		defer func() { _ = userRows.Close() }()

		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		for userRows.Next() {
			var configJSON []byte
			if err := userRows.Scan(&configJSON); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to scan user config_json: %w", err))
				mu.Unlock()
				return
			}

			var user configv1.User
			if err := opts.Unmarshal(configJSON, &user); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to unmarshal user config: %w", err))
				mu.Unlock()
				return
			}
			users = append(users, &user)
		}
		if err := userRows.Err(); err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to iterate user rows: %w", err))
			mu.Unlock()
		}
	}()

	// 3. Load Global Settings
	go func() {
		defer wg.Done()
		settingsRow := s.db.QueryRowContext(ctx, "SELECT config_json FROM global_settings WHERE id = 1")
		var settingsJSON []byte
		if err := settingsRow.Scan(&settingsJSON); err == nil {
			var s configv1.GlobalSettings
			opts := protojson.UnmarshalOptions{DiscardUnknown: true}
			if err := opts.Unmarshal(settingsJSON, &s); err == nil {
				settings = &s
			}
		}
	}()

	// 4. Load Collections
	go func() {
		defer wg.Done()
		collectionRows, err := s.db.QueryContext(ctx, "SELECT config_json FROM service_collections")
		if err != nil {
			// Ignore error as in original code
			return
		}
		defer func() { _ = collectionRows.Close() }()

		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
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
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to iterate collection rows: %w", err))
			mu.Unlock()
		}
	}()

	// 5. Load Profiles
	go func() {
		defer wg.Done()
		rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM profile_definitions")
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to query profile_definitions: %w", err))
			mu.Unlock()
			return
		}
		defer func() { _ = rows.Close() }()

		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		for rows.Next() {
			var configJSON []byte
			if err := rows.Scan(&configJSON); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to scan profile config_json: %w", err))
				mu.Unlock()
				return
			}
			var p configv1.ProfileDefinition
			if err := opts.Unmarshal(configJSON, &p); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to unmarshal profile config: %w", err))
				mu.Unlock()
				return
			}
			profiles = append(profiles, &p)
		}
		if err := rows.Err(); err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("failed to iterate profile rows: %w", err))
			mu.Unlock()
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	// Merge Profiles into Global Settings
	if len(profiles) > 0 {
		if settings == nil {
			settings = configv1.GlobalSettings_builder{}.Build()
		}
		current := settings.GetProfileDefinitions()
		current = append(current, profiles...)
		settings.SetProfileDefinitions(current)
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
// Summary: Persists an upstream service configuration to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - service: *configv1.UpstreamServiceConfig. The service configuration to save. Must have a valid name.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
//
// Throws/Errors:
//   - Returns error if "service name is required".
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

// GetService retrieves an upstream service configuration by name.
//
// Summary: Retrieves an upstream service configuration from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the service.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The requested service configuration if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves all upstream service configurations from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice of all service configurations.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Removes an upstream service configuration from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the service to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetGlobalSettings retrieves the global configuration.
//
// Summary: Retrieves the global settings from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - *configv1.GlobalSettings: The global settings object, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists the global settings to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - settings: *configv1.GlobalSettings. The global settings to save.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Creates a new user in PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - user: *configv1.User. The user object to create. Must have a valid ID.
//
// Returns:
//   - error: Returns nil on success, or an error if the user ID is missing or the operation fails.
//
// Throws/Errors:
//   - Returns error if "user ID is required".
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
// Summary: Retrieves a user from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - id: string. The unique identifier of the user.
//
// Returns:
//   - *configv1.User: The requested user if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves all users from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - []*configv1.User: A slice of all users.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Updates an existing user in PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - user: *configv1.User. The user object to update. Must correspond to an existing user.
//
// Returns:
//   - error: Returns nil on success, or an error if the user is not found or the operation fails.
//
// Throws/Errors:
//   - Returns error if "user ID is required".
//   - Returns error if "user not found".
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
// Summary: Removes a user from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - id: string. The unique identifier of the user to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves all secrets from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - []*configv1.Secret: A slice of all secrets.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves a specific secret from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - id: string. The unique identifier of the secret.
//
// Returns:
//   - *configv1.Secret: The requested secret if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists a secret to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - secret: *configv1.Secret. The secret to save. Must have a valid ID.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
//
// Throws/Errors:
//   - Returns error if "secret id is required".
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
// Summary: Removes a secret from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - id: string. The unique identifier of the secret to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves all profile definitions from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - []*configv1.ProfileDefinition: A slice of all profile definitions.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves a profile definition from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the profile.
//
// Returns:
//   - *configv1.ProfileDefinition: The requested profile if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists a profile definition to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - profile: *configv1.ProfileDefinition. The profile definition to save. Must have a valid name.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
//
// Throws/Errors:
//   - Returns error if "profile name is required".
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
// Summary: Removes a profile definition from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the profile to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves all service collections from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//
// Returns:
//   - []*configv1.Collection: A slice of all service collections.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Retrieves a service collection from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the collection.
//
// Returns:
//   - *configv1.Collection: The requested collection if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists a service collection to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - collection: *configv1.Collection. The service collection to save. Must have a valid name.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
//
// Throws/Errors:
//   - Returns error if "collection name is required".
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
// Summary: Removes a service collection from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - name: string. The unique name of the collection to delete.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Persists a user authentication token to PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - token: *configv1.UserToken. The token object to save. Must have user ID and service ID.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
//
// Throws/Errors:
//   - Returns error if "user ID and service ID are required".
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
// Summary: Retrieves a user authentication token from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *configv1.UserToken: The requested token if found, or nil if not found.
//   - error: Returns nil on success, or an error if the operation fails.
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
// Summary: Removes a user authentication token from PostgreSQL.
//
// Parameters:
//   - ctx: context.Context. Context for the operation.
//   - userID: string. The unique identifier of the user.
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - error: Returns nil on success, or an error if the operation fails.
func (s *Store) DeleteToken(ctx context.Context, userID, serviceID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = $1 AND service_id = $2", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
