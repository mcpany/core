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
	"golang.org/x/sync/errgroup"
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

	var paths []string
	// Walk the directory to collect file paths
	err := afero.Walk(m.fs, m.catalogPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Only look for config.yaml or popular/*.yaml
		if strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// ⚡ BOLT: Parallelize catalog loading to reduce startup time.
	// Randomized Selection from Top 5 High-Impact Targets
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	var sliceMu sync.Mutex

	for _, path := range paths {
		path := path
		g.Go(func() error {
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

			if services := cfg.GetUpstreamServices(); len(services) > 0 {
				sliceMu.Lock()
				m.services = append(m.services, services...)
				sliceMu.Unlock()
			}
			return nil
		})
	}

	return g.Wait()
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
