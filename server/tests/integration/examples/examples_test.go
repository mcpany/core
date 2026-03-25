// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package examples_test

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleConfigs(t *testing.T) {
	// Set dummy API key for validation to pass.
	t.Setenv("GEMINI_API_KEY", "dummy-key")
	projectRoot, rootErr := sourceProjectRoot()
	require.NoError(t, rootErr)
	runtimeRoot := filepath.Join(t.TempDir(), "server")
	examplesDir := filepath.Join(runtimeRoot, "examples")
	require.NoError(t, copyDir(filepath.Join(projectRoot, "examples"), examplesDir))

	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly.
	chdirErr := os.Chdir(runtimeRoot)
	require.NoError(t, chdirErr)

	// Ensure stdio example binary is built, as Config validation checks for its existence.
	stdioBinPath := filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, statErr := os.Stat(stdioBinPath); os.IsNotExist(statErr) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		cmd := exec.Command("go", "build", "-o", stdioBinPath, filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool", "main.go"))
		cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"), "GO111MODULE=off")
		cmd.Dir = runtimeRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if runErr := cmd.Run(); runErr != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", runErr)
		}
	}

	// Walk through examples directory.
	walkResultErr := filepath.Walk(examplesDir, func(path string, info os.FileInfo, walkDirErr error) error {
		if walkDirErr != nil {
			return walkDirErr
		}

		if !info.IsDir() && filepath.Base(path) == "config.yaml" {
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
	require.NoError(t, walkResultErr)
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

	store := config.NewFileStore(osFs, []string{configPath})
	configs, loadErr := config.LoadServices(context.Background(), store, "server")
	if loadErr != nil {
		t.Fatalf("Failed to load config %s: %v", configPath, loadErr)
	}

	validationErrors := config.Validate(context.Background(), configs, config.Server)
	assert.Empty(t, validationErrors, "Config validation failed for %s", configPath)
}
