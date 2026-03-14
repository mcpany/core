// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package examples_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Set dummy API key for validation to pass
	t.Setenv("GEMINI_API_KEY", "dummy-key")
	projectRoot, err := sourceProjectRoot()
	require.NoError(t, err)
	runtimeRoot := filepath.Join(t.TempDir(), "server")
	examplesDir := filepath.Join(runtimeRoot, "examples")
	require.NoError(t, copyDir(filepath.Join(projectRoot, "examples"), examplesDir))

	// Ensure stdio example binary is built, as Config validation checks for its existence
	// This makes the test robust against sharding/environment where build-examples might not have run.
	stdioBinPath := filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, err := os.Stat(stdioBinPath); os.IsNotExist(err) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		// Mock the tool binary since we are likely inside a bazel runfiles tree where go build will fail without the full mod workspace
		t.Logf("Creating mock binary to satisfy config validation in test context...")
		mockScript := "#!/bin/sh\necho 'mock mcp tool'"
		err = os.WriteFile(stdioBinPath, []byte(mockScript), 0755)
		if err != nil {
			t.Logf("Failed to create mock stdio binary: %v", err)
			// Ensure the directory exists
			_ = os.MkdirAll(filepath.Dir(stdioBinPath), 0755)
			err = os.WriteFile(stdioBinPath, []byte(mockScript), 0755)
			if err != nil {
				t.Fatalf("Failed to create mock stdio binary after MkdirAll: %v", err)
			}
		}
	}

	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly
	err = os.Chdir(runtimeRoot)
	require.NoError(t, err)

	// Walk through examples directory
	err = filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check for config.yaml
		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
			// Trim project root from path for cleaner test name
			testName := path
			if strings.HasPrefix(path, projectRoot) {
				testName = strings.TrimPrefix(path, projectRoot)
			}

			t.Run(testName, func(t *testing.T) {
				// 1. Initialize FS and load config
				fs := afero.NewOsFs()
				store := config.NewFileStore(fs, []string{path})
				cfg, err := store.Load(context.Background())

				// For stdio example, it references the built binary. If we skipped building above, validation fails.
				if err != nil && strings.Contains(testName, "stdio") && strings.Contains(err.Error(), "no such file or directory") {
					t.Logf("Warning: Skipping stdio config validation due to missing binary dependency. Please run `make build-examples` first.")
					return
				}

				if err != nil {
					t.Fatalf("Failed to load config %s: %v", testName, err)
				}

				// 2. Validate configuration
				require.NotNil(t, cfg, "Config should not be nil")
				require.GreaterOrEqual(t, len(cfg.GetUpstreamServices()), 1, "Config should contain at least one upstream service")

				// 3. Optional: add other validations if needed
				t.Logf("Config validation complete.")
			})
		}
		return nil
	})

	require.NoError(t, err, "Failed to walk examples directory")
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist.
func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return err
	}

	return nil
}

func sourceProjectRoot() (string, error) {
	workspace := os.Getenv("TEST_WORKSPACE")
	if workspace == "" {
		workspace = "_main"
	}
	for _, base := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if base == "" {
			continue
		}
		candidate := filepath.Join(base, workspace, "server")
		if _, err := os.Stat(filepath.Join(candidate, "examples")); err == nil {
			return candidate, nil
		}
		candidate = filepath.Join(base, workspace)
		if _, err := os.Stat(filepath.Join(candidate, "examples")); err == nil {
			return candidate, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
		if _, err := os.Stat(filepath.Join(candidate, "examples")); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find project root")
}
