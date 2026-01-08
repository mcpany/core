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

	// SaveService saves a single upstream service configuration.
	SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a single upstream service configuration by name.
	GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices lists all upstream service configurations.
	ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService deletes an upstream service configuration by name.
	DeleteService(ctx context.Context, name string) error

	// GetGlobalSettings retrieves the global configuration.
	GetGlobalSettings() (*configv1.GlobalSettings, error)

	// SaveGlobalSettings saves the global configuration.
	SaveGlobalSettings(settings *configv1.GlobalSettings) error

	// Secrets
	// ListSecrets retrieves all secrets.
	ListSecrets() ([]*configv1.Secret, error)
	// GetSecret retrieves a secret by ID.
	GetSecret(id string) (*configv1.Secret, error)
	// SaveSecret saves a secret.
	SaveSecret(secret *configv1.Secret) error
	// DeleteSecret deletes a secret by ID.
	DeleteSecret(id string) error

	// Users
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *configv1.User) error
	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, id string) (*configv1.User, error)
	// ListUsers retrieves all users.
	ListUsers(ctx context.Context) ([]*configv1.User, error)
	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, user *configv1.User) error
	// DeleteUser deletes a user by ID.
	DeleteUser(ctx context.Context, id string) error

	// Profiles
	// ListProfiles retrieves all profile definitions.
	ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error)
	// GetProfile retrieves a profile definition by name.
	GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error)
	// SaveProfile saves a profile definition.
	SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error
	// DeleteProfile deletes a profile definition by name.
	DeleteProfile(ctx context.Context, name string) error

	// Service Collections
	// ListServiceCollections retrieves all service collections.
	ListServiceCollections(ctx context.Context) ([]*configv1.UpstreamServiceCollectionShare, error)
	// GetServiceCollection retrieves a service collection by name.
	GetServiceCollection(ctx context.Context, name string) (*configv1.UpstreamServiceCollectionShare, error)
	// SaveServiceCollection saves a service collection.
	SaveServiceCollection(ctx context.Context, collection *configv1.UpstreamServiceCollectionShare) error
	// DeleteServiceCollection deletes a service collection by name.
	DeleteServiceCollection(ctx context.Context, name string) error

	// Close closes the underlying storage connection.
	Close() error
}
