// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
)

// Manager handles the loading and listing of catalog services.
//
// Summary: Manages the service catalog.
//
// It scans a specified directory for service configurations and provides access to them.
type Manager struct {
	mu          sync.RWMutex
	fs          afero.Fs
	catalogPath string
	services    []*configv1.UpstreamServiceConfig
}

// NewManager creates a new Catalog Manager.
//
// Summary: Initializes a new Catalog Manager.
//
// Parameters:
//   - fs: afero.Fs. The filesystem to scan.
//   - catalogPath: string. The path to the catalog directory.
//
// Returns:
//   - *Manager: The initialized manager.
func NewManager(fs afero.Fs, catalogPath string) *Manager {
	return &Manager{
		fs:          fs,
		catalogPath: catalogPath,
	}
}

// Load scans the catalog directory and loads all service configurations.
//
// Summary: Loads service configurations from the catalog directory.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - error: An error if the directory walk fails (individual config load errors are logged but do not abort).
//
// Side Effects:
//   - Updates the internal list of services.
//   - Reads files from the filesystem.
func (m *Manager) Load(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.services = nil // Reset

	// Walk the directory
	err := afero.Walk(m.fs, m.catalogPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Only look for config.yaml or popular/*.yaml
		// The moved structure is marketplace/catalog/<service_name>/config.yaml
		// OR we might have marketplace/upstream_service_collection/popular/*.yaml (which we created earlier)
		// Let's support both for now, or focus on the requested structure.
		// The user moved server/examples/popular_services/* -> marketplace/catalog/*
		// so we expect .../catalog/gemini/config.yaml etc.

		if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			// Load config
			store := config.NewFileStore(m.fs, []string{path})
			// We skip validation here to be lenient, or strict? Let's be strict but log errors.
			// Actually Store.Load returns McpAnyServerConfig
			cfg, loadErr := store.Load(ctx) // Renamed err to loadErr to avoid shadowing
			if loadErr != nil {
				// Log error but continue
				fmt.Printf("Failed to load catalog item %s: %v\n", path, loadErr)
				return nil
			}

			if cfg.GetUpstreamServices() != nil {
				m.services = append(m.services, cfg.GetUpstreamServices()...)
			}
		}
		return nil
	})

	return err
}

// ListServices returns the list of loaded services.
//
// Summary: Retrieves the list of loaded services.
//
// Parameters:
//   - _ context.Context: The context (unused).
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A slice of service configurations.
//   - error: Always nil.
func (m *Manager) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	result := make([]*configv1.UpstreamServiceConfig, len(m.services))
	copy(result, m.services)
	return result, nil
}
