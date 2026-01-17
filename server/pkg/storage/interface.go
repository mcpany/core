// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package storage defines the interface for persisting configuration.
package storage

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Storage defines the interface for persisting configuration.
type Storage interface {
	// Load retrieves the full server configuration.
	// Note: Currently it mostly returns UpstreamServices.
	// Load retrieves the full server configuration.
	// Note: Currently it mostly returns UpstreamServices.
	Load(ctx context.Context) (*configv1.McpAnyServerConfig, error)

	// HasConfigSources returns true if the store has configuration sources (e.g., file paths) configured.
	//
	// Returns true if successful.
	HasConfigSources() bool

	// SaveService saves a single upstream service configuration.
	//
	// ctx is the context for the request.
	// service is the service instance.
	//
	// Returns an error if the operation fails.
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a single upstream service configuration by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices lists all upstream service configurations.
	//
	// ctx is the context for the request.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService deletes an upstream service configuration by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns an error if the operation fails.
	DeleteService(ctx context.Context, name string) error

	// GetGlobalSettings retrieves the global configuration.
	//
	// ctx is the context for the request.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error)

	// SaveGlobalSettings saves the global configuration.
	//
	// ctx is the context for the request.
	// settings is the settings.
	//
	// Returns an error if the operation fails.
	SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error

	// Secrets
	// ListSecrets retrieves all secrets.
	ListSecrets(ctx context.Context) ([]*configv1.Secret, error)
	// GetSecret retrieves a secret by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetSecret(ctx context.Context, id string) (*configv1.Secret, error)
	// SaveSecret saves a secret.
	//
	// ctx is the context for the request.
	// secret is the secret.
	//
	// Returns an error if the operation fails.
	SaveSecret(ctx context.Context, secret *configv1.Secret) error
	// DeleteSecret deletes a secret by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns an error if the operation fails.
	DeleteSecret(ctx context.Context, id string) error

	// Users
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *configv1.User) error
	// GetUser retrieves a user by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetUser(ctx context.Context, id string) (*configv1.User, error)
	// ListUsers retrieves all users.
	//
	// ctx is the context for the request.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	ListUsers(ctx context.Context) ([]*configv1.User, error)
	// UpdateUser updates an existing user.
	//
	// ctx is the context for the request.
	// user is the user.
	//
	// Returns an error if the operation fails.
	UpdateUser(ctx context.Context, user *configv1.User) error
	// DeleteUser deletes a user by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns an error if the operation fails.
	DeleteUser(ctx context.Context, id string) error

	// Profiles
	// ListProfiles retrieves all profile definitions.
	ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error)
	// GetProfile retrieves a profile definition by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error)
	// SaveProfile saves a profile definition.
	//
	// ctx is the context for the request.
	// profile is the profile.
	//
	// Returns an error if the operation fails.
	SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error
	// DeleteProfile deletes a profile definition by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns an error if the operation fails.
	DeleteProfile(ctx context.Context, name string) error

	// Service Collections
	// ListServiceCollections retrieves all service collections.
	ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error)
	// GetServiceCollection retrieves a service collection by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error)
	// SaveServiceCollection saves a service collection.
	//
	// ctx is the context for the request.
	// collection is the collection.
	//
	// Returns an error if the operation fails.
	SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error
	// DeleteServiceCollection deletes a service collection by name.
	//
	// ctx is the context for the request.
	// name is the name of the resource.
	//
	// Returns an error if the operation fails.
	DeleteServiceCollection(ctx context.Context, name string) error

	// Tokens
	// SaveToken saves a user token.
	SaveToken(ctx context.Context, token *configv1.UserToken) error
	// GetToken retrieves a user token by user ID and service ID.
	//
	// ctx is the context for the request.
	// userID is the userID.
	// serviceID is the serviceID.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error)
	// DeleteToken deletes a user token.
	//
	// ctx is the context for the request.
	// userID is the userID.
	// serviceID is the serviceID.
	//
	// Returns an error if the operation fails.
	DeleteToken(ctx context.Context, userID, serviceID string) error

	// Credentials
	// ListCredentials retrieves all credentials.
	ListCredentials(ctx context.Context) ([]*configv1.Credential, error)
	// GetCredential retrieves a credential by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	GetCredential(ctx context.Context, id string) (*configv1.Credential, error)
	// SaveCredential saves a credential.
	//
	// ctx is the context for the request.
	// cred is the cred.
	//
	// Returns an error if the operation fails.
	SaveCredential(ctx context.Context, cred *configv1.Credential) error
	// DeleteCredential deletes a credential by ID.
	//
	// ctx is the context for the request.
	// id is the unique identifier.
	//
	// Returns an error if the operation fails.
	DeleteCredential(ctx context.Context, id string) error

	// Close closes the underlying storage connection.
	//
	// Returns an error if the operation fails.
	Close() error
}
