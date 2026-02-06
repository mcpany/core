// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package storage defines the interface for persisting configuration.
package storage

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

// Storage defines the interface for persisting configuration.
//
// Summary: Interface for backend storage operations.
type Storage interface {
	// Logs
	// QueryLogs retrieves logs based on filters.
	QueryLogs(ctx context.Context, filter logging.LogFilter) ([]logging.LogEntry, int, error)

	// Load retrieves the full server configuration.
	//
	// Summary: Loads the entire server configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *configv1.McpAnyServerConfig: The loaded configuration.
	//   - error: An error if loading fails.
	Load(ctx context.Context) (*configv1.McpAnyServerConfig, error)

	// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
	//
	// Summary: Checks if the store has any configuration sources.
	//
	// Returns:
	//   - bool: True if sources exist.
	HasConfigSources() bool

	// SaveService saves a single upstream service configuration.
	//
	// Summary: Persists a service configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - service: *configv1.UpstreamServiceConfig. The service configuration.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a single upstream service configuration by name.
	//
	// Summary: Retrieves a service configuration by name.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The name of the service.
	//
	// Returns:
	//   - *configv1.UpstreamServiceConfig: The service configuration.
	//   - error: An error if retrieval fails.
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices lists all upstream service configurations.
	//
	// Summary: Lists all services.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.UpstreamServiceConfig: A list of service configurations.
	//   - error: An error if listing fails.
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService deletes an upstream service configuration by name.
	//
	// Summary: Deletes a service configuration.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The name of the service to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteService(ctx context.Context, name string) error

	// GetGlobalSettings retrieves the global configuration.
	//
	// Summary: Retrieves global settings.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - *configv1.GlobalSettings: The global settings.
	//   - error: An error if retrieval fails.
	GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error)

	// SaveGlobalSettings saves the global configuration.
	//
	// Summary: Persists global settings.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - settings: *configv1.GlobalSettings. The settings to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error

	// ListSecrets retrieves all secrets.
	//
	// Summary: Lists all secrets.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.Secret: A list of secrets.
	//   - error: An error if listing fails.
	ListSecrets(ctx context.Context) ([]*configv1.Secret, error)

	// GetSecret retrieves a secret by ID.
	//
	// Summary: Retrieves a secret by ID.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The secret ID.
	//
	// Returns:
	//   - *configv1.Secret: The secret.
	//   - error: An error if retrieval fails.
	GetSecret(ctx context.Context, id string) (*configv1.Secret, error)

	// SaveSecret saves a secret.
	//
	// Summary: Persists a secret.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - secret: *configv1.Secret. The secret to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveSecret(ctx context.Context, secret *configv1.Secret) error

	// DeleteSecret deletes a secret by ID.
	//
	// Summary: Deletes a secret.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The secret ID to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteSecret(ctx context.Context, id string) error

	// CreateUser creates a new user.
	//
	// Summary: Creates a user.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - user: *configv1.User. The user to create.
	//
	// Returns:
	//   - error: An error if creation fails.
	CreateUser(ctx context.Context, user *configv1.User) error

	// GetUser retrieves a user by ID.
	//
	// Summary: Retrieves a user by ID.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The user ID.
	//
	// Returns:
	//   - *configv1.User: The user.
	//   - error: An error if retrieval fails.
	GetUser(ctx context.Context, id string) (*configv1.User, error)

	// ListUsers retrieves all users.
	//
	// Summary: Lists all users.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.User: A list of users.
	//   - error: An error if listing fails.
	ListUsers(ctx context.Context) ([]*configv1.User, error)

	// UpdateUser updates an existing user.
	//
	// Summary: Updates a user.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - user: *configv1.User. The user to update.
	//
	// Returns:
	//   - error: An error if update fails.
	UpdateUser(ctx context.Context, user *configv1.User) error

	// DeleteUser deletes a user by ID.
	//
	// Summary: Deletes a user.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The user ID to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteUser(ctx context.Context, id string) error

	// ListProfiles retrieves all profile definitions.
	//
	// Summary: Lists all profiles.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.ProfileDefinition: A list of profiles.
	//   - error: An error if listing fails.
	ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error)

	// GetProfile retrieves a profile definition by name.
	//
	// Summary: Retrieves a profile by name.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The profile name.
	//
	// Returns:
	//   - *configv1.ProfileDefinition: The profile.
	//   - error: An error if retrieval fails.
	GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error)

	// SaveProfile saves a profile definition.
	//
	// Summary: Persists a profile.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - profile: *configv1.ProfileDefinition. The profile to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error

	// DeleteProfile deletes a profile definition by name.
	//
	// Summary: Deletes a profile.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The profile name to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteProfile(ctx context.Context, name string) error

	// ListServiceCollections retrieves all service collections.
	//
	// Summary: Lists all collections.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.Collection: A list of collections.
	//   - error: An error if listing fails.
	ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error)

	// GetServiceCollection retrieves a service collection by name.
	//
	// Summary: Retrieves a collection by name.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The collection name.
	//
	// Returns:
	//   - *configv1.Collection: The collection.
	//   - error: An error if retrieval fails.
	GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error)

	// SaveServiceCollection saves a service collection.
	//
	// Summary: Persists a collection.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - collection: *configv1.Collection. The collection to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error

	// DeleteServiceCollection deletes a service collection by name.
	//
	// Summary: Deletes a collection.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - name: string. The collection name to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteServiceCollection(ctx context.Context, name string) error

	// SaveToken saves a user token.
	//
	// Summary: Persists a user token.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - token: *configv1.UserToken. The token to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveToken(ctx context.Context, token *configv1.UserToken) error

	// GetToken retrieves a user token by user ID and service ID.
	//
	// Summary: Retrieves a user token.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - userID: string. The user ID.
	//   - serviceID: string. The service ID.
	//
	// Returns:
	//   - *configv1.UserToken: The token.
	//   - error: An error if retrieval fails.
	GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error)

	// DeleteToken deletes a user token.
	//
	// Summary: Deletes a user token.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - userID: string. The user ID.
	//   - serviceID: string. The service ID.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteToken(ctx context.Context, userID, serviceID string) error

	// ListCredentials retrieves all credentials.
	//
	// Summary: Lists all credentials.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//
	// Returns:
	//   - []*configv1.Credential: A list of credentials.
	//   - error: An error if listing fails.
	ListCredentials(ctx context.Context) ([]*configv1.Credential, error)

	// GetCredential retrieves a credential by ID.
	//
	// Summary: Retrieves a credential by ID.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The credential ID.
	//
	// Returns:
	//   - *configv1.Credential: The credential.
	//   - error: An error if retrieval fails.
	GetCredential(ctx context.Context, id string) (*configv1.Credential, error)

	// SaveCredential saves a credential.
	//
	// Summary: Persists a credential.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - cred: *configv1.Credential. The credential to save.
	//
	// Returns:
	//   - error: An error if saving fails.
	SaveCredential(ctx context.Context, cred *configv1.Credential) error

	// DeleteCredential deletes a credential by ID.
	//
	// Summary: Deletes a credential.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - id: string. The credential ID to delete.
	//
	// Returns:
	//   - error: An error if deletion fails.
	DeleteCredential(ctx context.Context, id string) error

	// Close closes the underlying storage connection.
	//
	// Summary: Closes the storage connection.
	//
	// Returns:
	//   - error: An error if closing fails.
	Close() error
}
