// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
)

// Store represents the public Store entity.
//
// Summary: Defines the structured data model representing a .
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
	db *DB
}

// NewStore serves as a public interface for interacting with NewStore.
//
// Summary: Constructs and returns an initialized store ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewStore(db *DB) *Store {
	return &Store{db: db}
}

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) Close() error {
	return s.db.Close()
}

// HasConfigSources serves as a public interface for interacting with HasConfigSources.
//
// Summary: Checks condition indicating whether the target has config sources.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) HasConfigSources() bool {
	return true
}

// Load serves as a public interface for interacting with Load.
//
// Summary: Load the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SaveService serves as a public interface for interacting with SaveService.
//
// Summary: Save the service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetService serves as a public interface for interacting with GetService.
//
// Summary: Fetches and returns the underlying service from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ListServices serves as a public interface for interacting with ListServices.
//
// Summary: List the services appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteService serves as a public interface for interacting with DeleteService.
//
// Summary: Delete the service appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteService(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM upstream_services WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

// GetGlobalSettings serves as a public interface for interacting with GetGlobalSettings.
//
// Summary: Fetches and returns the underlying global settings from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SaveGlobalSettings serves as a public interface for interacting with SaveGlobalSettings.
//
// Summary: Save the global settings appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// CreateUser serves as a public interface for interacting with CreateUser.
//
// Summary: Create the user appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetUser serves as a public interface for interacting with GetUser.
//
// Summary: Fetches and returns the underlying user from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ListUsers serves as a public interface for interacting with ListUsers.
//
// Summary: List the users appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// UpdateUser serves as a public interface for interacting with UpdateUser.
//
// Summary: Update the user appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteUser serves as a public interface for interacting with DeleteUser.
//
// Summary: Delete the user appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteUser(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Secrets

// ListSecrets serves as a public interface for interacting with ListSecrets.
//
// Summary: List the secrets appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetSecret serves as a public interface for interacting with GetSecret.
//
// Summary: Fetches and returns the underlying secret from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SaveSecret serves as a public interface for interacting with SaveSecret.
//
// Summary: Save the secret appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteSecret serves as a public interface for interacting with DeleteSecret.
//
// Summary: Delete the secret appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteSecret(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM secrets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// SaveLog serves as a public interface for interacting with SaveLog.
//
// Summary: Save the log appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) SaveLog(ctx context.Context, entry *logging.LogEntry) error {
	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
	INSERT INTO logs (id, timestamp, level, source, message, metadata_json, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	ON CONFLICT (id) DO NOTHING
	`
	_, err = s.db.ExecContext(ctx, query, entry.ID, entry.Timestamp, entry.Level, entry.Source, entry.Message, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to save log: %w", err)
	}
	return nil
}

// GetRecentLogs serves as a public interface for interacting with GetRecentLogs.
//
// Summary: Fetches and returns the underlying recent logs from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) GetRecentLogs(ctx context.Context, limit int) ([]*logging.LogEntry, error) {
	query := `
    SELECT id, timestamp, level, source, message, metadata_json
    FROM (
        SELECT id, timestamp, level, source, message, metadata_json
        FROM logs
        ORDER BY timestamp DESC
        LIMIT $1
    ) AS recent_logs
    ORDER BY timestamp ASC
    `
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []*logging.LogEntry
	for rows.Next() {
		var entry logging.LogEntry
		var metadataJSON string
		var ts time.Time
		var source sql.NullString

		if err := rows.Scan(&entry.ID, &ts, &entry.Level, &source, &entry.Message, &metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		entry.Timestamp = ts.Format(time.RFC3339)
		if source.Valid {
			entry.Source = source.String
		}

		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &entry.Metadata); err != nil {
				// Ignore unmarshal error
				_ = err
			}
		}
		logs = append(logs, &entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating logs: %w", err)
	}
	return logs, nil
}

// Profiles

// ListProfiles serves as a public interface for interacting with ListProfiles.
//
// Summary: List the profiles appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetProfile serves as a public interface for interacting with GetProfile.
//
// Summary: Fetches and returns the underlying profile from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SaveProfile serves as a public interface for interacting with SaveProfile.
//
// Summary: Save the profile appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteProfile serves as a public interface for interacting with DeleteProfile.
//
// Summary: Delete the profile appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteProfile(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM profile_definitions WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	return nil
}

// Service Collections

// ListServiceCollections serves as a public interface for interacting with ListServiceCollections.
//
// Summary: List the service collections appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetServiceCollection serves as a public interface for interacting with GetServiceCollection.
//
// Summary: Fetches and returns the underlying service collection from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// SaveServiceCollection serves as a public interface for interacting with SaveServiceCollection.
//
// Summary: Save the service collection appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteServiceCollection serves as a public interface for interacting with DeleteServiceCollection.
//
// Summary: Delete the service collection appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteServiceCollection(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM service_collections WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}
	return nil
}

// Tokens

// SaveToken serves as a public interface for interacting with SaveToken.
//
// Summary: Save the token appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetToken serves as a public interface for interacting with GetToken.
//
// Summary: Fetches and returns the underlying token from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// DeleteToken serves as a public interface for interacting with DeleteToken.
//
// Summary: Delete the token appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteToken(ctx context.Context, userID, serviceID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM user_tokens WHERE user_id = $1 AND service_id = $2", userID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// Credentials

// ListCredentials serves as a public interface for interacting with ListCredentials.
//
// Summary: List the credentials appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// GetCredential serves as a public interface for interacting with GetCredential.
//
// Summary: Fetches and returns the underlying credential from the system state.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	query := "SELECT config_json FROM credentials WHERE id = $1"
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

// SaveCredential serves as a public interface for interacting with SaveCredential.
//
// Summary: Save the credential appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
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

// DeleteCredential serves as a public interface for interacting with DeleteCredential.
//
// Summary: Delete the credential appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (s *Store) DeleteCredential(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM credentials WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}
