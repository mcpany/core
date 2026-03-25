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

// Manager handles the loading and listing of catalog services.
// NewManager creates a new Catalog Manager.
//
// Summary: Initializes a new Catalog Manager.
// Load scans the catalog directory and loads all service configurations.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Summary: Loads service configurations from the catalog directory.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Returns:
//   - error: An error if the directory walk fails (individual config load errors are logged but do not abort).
//
// Side Effects:
//   - Updates the internal list of services.
//   - Reads files from the filesystem.
// Errors:
//   - triggers relevant error states on failure.
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
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *Manager) ListServices(_ context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	result := make([]*configv1.UpstreamServiceConfig, len(m.services))
	copy(result, m.services)
	return result, nil
}
