// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Storage defines the interface for persisting configuration.
type Storage interface {
	// Load retrieves the full server configuration.
	// Note: Currently it mostly returns UpstreamServices.
	Load() (*configv1.McpAnyServerConfig, error)

	// SaveService saves a single upstream service configuration.
	SaveService(service *configv1.UpstreamServiceConfig) error

	// GetService retrieves a single upstream service configuration by name.
	GetService(name string) (*configv1.UpstreamServiceConfig, error)

	// ListServices lists all upstream service configurations.
	ListServices() ([]*configv1.UpstreamServiceConfig, error)

	// DeleteService deletes an upstream service configuration by name.
	DeleteService(name string) error

	// Close closes the underlying storage connection.
	Close() error
}
