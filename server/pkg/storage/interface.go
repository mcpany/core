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

	// Close closes the underlying storage connection.
	Close() error
}
