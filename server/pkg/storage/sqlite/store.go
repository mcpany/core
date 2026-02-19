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
	db *DB
}

// NewStore initializes a new SQLite storage backend.
//
// Summary: Creates a SQLite store.
//
// Parameters:
//   - db: *DB. The SQLite database connection.
//
// Returns:
//   - *Store: The initialized store.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// Close closes the underlying SQLite database connection.
//
// Summary: Closes the store.
//
// Returns:
//   - error: An error if closing the database fails.
func (s *Store) Close() error {
	return s.db.Close()
}

// HasConfigSources checks if the store has configuration sources.
//
// Summary: Checks for configuration sources.
//
// Returns:
//   - bool: Always true for database stores.
func (s *Store) HasConfigSources() bool {
	return true
}

// Load retrieves the complete server configuration from the database.
//
// Summary: Loads server configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - *configv1.McpAnyServerConfig: The loaded configuration.
//   - error: An error if loading fails.
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

	// ⚡ BOLT: Optimized memory allocation by scanning directly into []byte.
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

// SaveService persists an upstream service configuration.
//
// Summary: Saves a service.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - service: *configv1.UpstreamServiceConfig. The service configuration to save.
//
// Returns:
//   - error: An error if the save operation fails.
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

// GetService retrieves a service configuration by its name.
//
// Summary: Retrieves a service.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - name: string. The name of the service.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The service configuration, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	query := "SELECT config_json FROM upstream_services WHERE name = ?"
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

// ListServices retrieves all configured upstream services.
//
// Summary: Lists all services.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of service configurations.
//   - error: An error if the query fails.
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

// DeleteService removes an upstream service configuration.
//
// Summary: Deletes a service.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - name: string. The name of the service to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetGlobalSettings retrieves the global server settings.
//
// Summary: Retrieves global settings.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - *configv1.GlobalSettings: The global settings configuration.
//   - error: An error if the query fails.
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

// SaveGlobalSettings persists the global server settings.
//
// Summary: Saves global settings.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - settings: *configv1.GlobalSettings. The settings to save.
//
// Returns:
//   - error: An error if the save operation fails.
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
	return nil
}

// Users

// CreateUser persists a new user configuration.
//
// Summary: Creates a user.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - user: *configv1.User. The user configuration to create.
//
// Returns:
//   - error: An error if the creation fails.
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

// GetUser retrieves a user configuration by ID.
//
// Summary: Retrieves a user.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - id: string. The user ID.
//
// Returns:
//   - *configv1.User: The user configuration, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	query := "SELECT config_json FROM users WHERE id = ?"
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

// ListUsers retrieves all configured users.
//
// Summary: Lists all users.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.User: A list of user configurations.
//   - error: An error if the query fails.
func (s *Store) ListUsers(ctx context.Context) ([]*configv1.User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []*configv1.User
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan user config: %w", err)
		}

		var user configv1.User
		if err := protojson.Unmarshal(configJSON, &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user config: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// UpdateUser updates an existing user configuration.
//
// Summary: Updates a user.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - user: *configv1.User. The updated user configuration.
//
// Returns:
//   - error: An error if the update fails or the user is not found.
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

// DeleteUser removes a user configuration by ID.
//
// Summary: Deletes a user.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - id: string. The ID of the user to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Secrets

// ListSecrets retrieves all configured secrets.
//
// Summary: Lists all secrets.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.Secret: A list of secrets.
//   - error: An error if the query fails.
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

// GetSecret retrieves a secret by its ID.
//
// Summary: Retrieves a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - id: string. The secret ID.
//
// Returns:
//   - *configv1.Secret: The secret, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	query := "SELECT config_json FROM secrets WHERE id = ?"
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

// SaveSecret persists a secret configuration.
//
// Summary: Saves a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - secret: *configv1.Secret. The secret to save.
//
// Returns:
//   - error: An error if the save operation fails.
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
	return nil
}

// DeleteSecret removes a secret by ID.
//
// Summary: Deletes a secret.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - id: string. The ID of the secret to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteSecret(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM secrets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// Profiles

// ListProfiles retrieves all profile definitions.
//
// Summary: Lists all profiles.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.ProfileDefinition: A list of profile definitions.
//   - error: An error if the query fails.
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
// Summary: Retrieves a profile.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - name: string. The name of the profile.
//
// Returns:
//   - *configv1.ProfileDefinition: The profile definition, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	query := "SELECT config_json FROM profile_definitions WHERE name = ?"
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

// SaveProfile persists a profile definition.
//
// Summary: Saves a profile.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - profile: *configv1.ProfileDefinition. The profile to save.
//
// Returns:
//   - error: An error if the save operation fails.
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

// DeleteProfile removes a profile definition by name.
//
// Summary: Deletes a profile.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - name: string. The name of the profile to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM profile_definitions WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// Service Collections

// ListServiceCollections retrieves all service collections.
//
// Summary: Lists all collections.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.Collection: A list of service collections.
//   - error: An error if the query fails.
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
// Summary: Retrieves a collection.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - name: string. The name of the collection.
//
// Returns:
//   - *configv1.Collection: The collection, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	query := "SELECT config_json FROM service_collections WHERE name = ?"
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

// SaveServiceCollection persists a service collection.
//
// Summary: Saves a collection.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - collection: *configv1.Collection. The collection to save.
//
// Returns:
//   - error: An error if the save operation fails.
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
	return nil
}

// DeleteServiceCollection removes a service collection by name.
//
// Summary: Deletes a collection.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - name: string. The name of the collection to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteServiceCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_collections WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

// Tokens

// SaveToken persists a user token.
//
// Summary: Saves a user token.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - token: *configv1.UserToken. The token to save.
//
// Returns:
//   - error: An error if the save operation fails.
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

// GetToken retrieves a user token.
//
// Summary: Retrieves a token.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - userID: string. The user ID.
//   - serviceID: string. The service ID.
//
// Returns:
//   - *configv1.UserToken: The token, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	query := "SELECT config_json FROM user_tokens WHERE user_id = ? AND service_id = ?"
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

// DeleteToken removes a user token.
//
// Summary: Deletes a token.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - userID: string. The user ID.
//   - serviceID: string. The service ID.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteToken(ctx context.Context, userID, serviceID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = ? AND service_id = ?", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// Credentials

// ListCredentials retrieves all configured credentials.
//
// Summary: Lists all credentials.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//
// Returns:
//   - []*configv1.Credential: A list of credentials.
//   - error: An error if the query fails.
func (s *Store) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT config_json FROM credentials")
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var credentials []*configv1.Credential
	for rows.Next() {
		var configJSON []byte
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan credential config: %w", err)
		}

		var cred configv1.Credential
		if err := protojson.Unmarshal(configJSON, &cred); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
		}
		credentials = append(credentials, &cred)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return credentials, nil
}

// GetCredential retrieves a credential by ID.
//
// Summary: Retrieves a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the database query.
//   - id: string. The credential ID.
//
// Returns:
//   - *configv1.Credential: The credential, or nil if not found.
//   - error: An error if the query fails.
func (s *Store) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	query := "SELECT config_json FROM credentials WHERE id = ?"
	row := s.db.QueryRowContext(ctx, query, id)

	var configJSON []byte
	if err := row.Scan(&configJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	var cred configv1.Credential
	if err := protojson.Unmarshal(configJSON, &cred); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return &cred, nil
}

// SaveCredential persists a credential.
//
// Summary: Saves a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - cred: *configv1.Credential. The credential to save.
//
// Returns:
//   - error: An error if the save operation fails.
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
	return nil
}

// DeleteCredential removes a credential by ID.
//
// Summary: Deletes a credential.
//
// Parameters:
//   - ctx: context.Context. The context for the database operation.
//   - id: string. The credential ID to delete.
//
// Returns:
//   - error: An error if the deletion fails.
func (s *Store) DeleteCredential(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}
