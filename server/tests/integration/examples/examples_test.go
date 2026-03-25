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
	projectRoot, rootFetchErr := sourceProjectRoot()
	require.NoError(t, rootFetchErr)
	runtimeRoot := filepath.Join(t.TempDir(), "server")
	examplesDir := filepath.Join(runtimeRoot, "examples")
	require.NoError(t, copyDir(filepath.Join(projectRoot, "examples"), examplesDir))

	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly.
	chdirErr := os.Chdir(runtimeRoot)
	require.NoError(t, chdirErr)

	// Ensure stdio example binary is built, as Config validation checks for its existence.
	stdioBinPath := filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool-bin")
	if _, binStatErr := os.Stat(stdioBinPath); os.IsNotExist(binStatErr) {
		t.Logf("Building missing stdio example binary: %s", stdioBinPath)
		buildCmd := exec.Command("go", "build", "-o", stdioBinPath, filepath.Join(runtimeRoot, "examples", "demo", "stdio", "my-tool", "main.go"))
		buildCmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"), "GO111MODULE=off")
		buildCmd.Dir = runtimeRoot
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if buildRunErr := buildCmd.Run(); buildRunErr != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", buildRunErr)
		}
	}

	// Walk through examples directory.
	walkResultErr := filepath.Walk(examplesDir, func(examplePath string, exampleInfo os.FileInfo, walkDirErr error) error {
		if walkDirErr != nil {
			return walkDirErr
		}

		if !exampleInfo.IsDir() && filepath.Base(examplePath) == "config.yaml" {
			testCaseName := examplePath
			if strings.HasPrefix(examplePath, projectRoot) {
				testCaseName = strings.TrimPrefix(examplePath, projectRoot)
			}
			t.Run(testCaseName, func(subT *testing.T) {
				validateConfig(subT, examplePath)
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
	for _, searchBase := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if searchBase == "" {
			continue
		}
		candidatePath := filepath.Join(searchBase, workspace, "server")
		if _, candidateStatErr := os.Stat(filepath.Join(candidatePath, "examples")); candidateStatErr == nil {
			return candidatePath, nil
		}
	}
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", os.ErrNotExist
	}
	return filepath.Abs(filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../..")))
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(itemPath string, itemInfo os.FileInfo, copyWalkErr error) error {
		if copyWalkErr != nil {
			return copyWalkErr
		}
		itemRelPath, relErr := filepath.Rel(src, itemPath)
		if relErr != nil {
			return relErr
		}
		targetPath := filepath.Join(dst, itemRelPath)
		if itemInfo.IsDir() {
			return os.MkdirAll(targetPath, itemInfo.Mode())
		}
		inputFile, openErr := os.Open(itemPath)
		if openErr != nil {
			return openErr
		}
		defer inputFile.Close()
		outputFile, createErr := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, itemInfo.Mode())
		if createErr != nil {
			return createErr
		}
		defer outputFile.Close()
		_, copyIoErr := io.Copy(outputFile, inputFile)
		return copyIoErr
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

	for _, envVarName := range requiredEnvVars {
		t.Setenv(envVarName, "dummy-val")
	}

	configStore := config.NewFileStore(osFs, []string{configPath})
	serviceConfigs, configLoadErr := config.LoadServices(context.Background(), configStore, "server")
	if configLoadErr != nil {
		t.Fatalf("Failed to load config %s: %v", configPath, configLoadErr)
	}

	validationErrors := config.Validate(context.Background(), serviceConfigs, config.Server)
	assert.Empty(t, validationErrors, "Config validation failed for %s", configPath)
}
