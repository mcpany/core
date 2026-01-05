// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package examples_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Set dummy API key for validation to pass
	t.Setenv("GEMINI_API_KEY", "dummy-key")
	// Find the project root
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Assuming we run this from anywhere in the repo, we need to find the root.
	// Common pattern: look for go.mod
	projectRoot := findProjectRoot(t, wd)
	examplesDir := filepath.Join(projectRoot, "examples")

	// Walk through examples directory
	err = filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for config.yaml
		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
			t.Run(path, func(t *testing.T) {
				validateConfig(t, path)
			})
		}
		return nil
	})
	require.NoError(t, err)
}

func validateConfig(t *testing.T, configPath string) {
	osFs := afero.NewOsFs()

	// Create a store that points to this config file
	store := config.NewFileStore(osFs, []string{configPath})

	// Load services
	// The second argument "server" matches what the CLI uses for validation context if any
	configs, err := config.LoadServices(context.Background(), store, "server")
	if err != nil {
		// Some configs might require env vars which validly fail if missing.
		// However, LoadServices typically parses the YAML/Proto.
		// If it fails due to missing env vars that are required for *parsing* (if any), that might be acceptable if we can detect it.
		// But usually configs placeholders are just strings unless they are used in a way that breaks parsing.
		// Let's see if we fail on basic loading.
		t.Fatalf("Failed to load config %s: %v", configPath, err)
	}

	// Validate
	validationErrors := config.Validate(context.Background(), configs, config.Server)
	assert.Empty(t, validationErrors, "Config validation failed for %s", configPath)
}

func findProjectRoot(t *testing.T, startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod)")
		}
		dir = parent
	}
}
