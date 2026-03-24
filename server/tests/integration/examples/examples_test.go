// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package examples_test

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
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

	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly
	err = os.Chdir(runtimeRoot)
	require.NoError(t, err)

	// Ensure stdio example binary is built, as Config validation checks for its existence
	// This makes the test robust against sharding/environment where build-examples might not have run.
	stdioBinPath := filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, statErr := os.Stat(stdioBinPath); os.IsNotExist(statErr) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		// Set GO111MODULE=off because we are building a simple main.go in a temp dir without a go.mod.
		cmd := exec.Command("go", "build", "-o", stdioBinPath, filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool", "main.go"))
		cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"), "GO111MODULE=off")
		cmd.Dir = runtimeRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if runErr := cmd.Run(); runErr != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", runErr)
		}
	}

	// Walk through examples directory
	err = filepath.Walk(examplesDir, func(path string, info os.FileInfo, walkDirErr error) error {
		if walkDirErr != nil {
			return walkDirErr
		}

		// Check for config.yaml
		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
			// Trim project root from path for cleaner test name
			testName := path
			if strings.HasPrefix(path, projectRoot) {
				testName = strings.TrimPrefix(path, projectRoot)
			}
			t.Run(testName, func(t *testing.T) {
				validateConfig(t, path)
			})
		}
		return nil
	})
	require.NoError(t, err)
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
		if _, statErr := os.Stat(filepath.Join(candidate, "examples")); statErr == nil {
			return candidate, nil
		}
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	return filepath.Abs(filepath.Clean(filepath.Join(filepath.Dir(file), "../../..")))
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, copyWalkErr error) error {
		if copyWalkErr != nil {
			return copyWalkErr
		}
		relPath, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		target := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		in, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		defer in.Close()
		out, createErr := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if createErr != nil {
			return createErr
		}
		defer out.Close()
		_, copyErr := io.Copy(out, in)
		return copyErr
	})
}

func validateConfig(t *testing.T, configPath string) {
	osFs := afero.NewOsFs()

	// Set dummy values for all required environment variables found in failure logs
	// This allows the strict config validation to pass during tests
	requiredEnvVars := []string{
		"AIRTABLE_API_TOKEN",
		"FIGMA_API_TOKEN",
		"GITHUB_TOKEN",
		"GOOGLE_API_KEY",
		"GOOGLE_OAUTH_CLIENT_ID",
		"GOOGLE_OAUTH_CLIENT_SECRET",
		"GOOGLE_OAUTH_REFRESH_TOKEN",
		"IPINFO_API_TOKEN",
		"MIRO_API_TOKEN",
		"NASA_OPEN_API_KEY",
		"SLACK_API_TOKEN",
		"STRIPE_API_KEY",
		"TRELLO_API_TOKEN",
		"TRELLO_API_KEY",
		"TWILIO_ACCOUNT_SID",
		"TWILIO_API_KEY",
		"TWILIO_API_SECRET",
	}

	for _, v := range requiredEnvVars {
		t.Setenv(v, "dummy-val")
	}

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
