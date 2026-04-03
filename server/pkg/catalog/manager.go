// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
)

// Manager represents the public Manager entity.
//
// Summary: Coordinates operations and orchestrates lifecycle events for the  components.
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
type Manager struct {
	mu          sync.RWMutex
	fs          afero.Fs
	catalogPath string
	services    []*configv1.UpstreamServiceConfig
}

// NewManager serves as a public interface for interacting with NewManager.
//
// Summary: Constructs and returns an initialized manager ready for consumption.
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
func NewManager(fs afero.Fs, catalogPath string) *Manager {
	return &Manager{
		fs:          fs,
		catalogPath: catalogPath,
	}
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
func (m *Manager) Load(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.services = nil // Reset

	var paths []string

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
	// Limit concurrency to avoid overwhelming the system (e.g. file descriptors)
	g.SetLimit(runtime.GOMAXPROCS(0) * 2)

	var mu sync.Mutex
	m.services = make([]*configv1.UpstreamServiceConfig, 0, len(paths))

	for _, path := range paths {
		path := path // Capture loop variable
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

			if services := cfg.GetUpstreamServices(); services != nil {
				mu.Lock()
				m.services = append(m.services, services...)
				mu.Unlock()
			}
			return nil
		})
	}

	return g.Wait()
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
func (m *Manager) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	result := make([]*configv1.UpstreamServiceConfig, len(m.services))
	copy(result, m.services)
	return result, nil
}
