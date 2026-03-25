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
	projectRootPath, rootFetchErr := sourceProjectRoot()
	require.NoError(t, rootFetchErr)
	runtimeRootPath := filepath.Join(t.TempDir(), "server")
	examplesDirectory := filepath.Join(runtimeRootPath, "examples")
	require.NoError(t, copyDir(filepath.Join(projectRootPath, "examples"), examplesDirectory))

	// Change to project root so that relative paths in configs (e.g. "./examples/...") resolve correctly.
	dirChangeErr := os.Chdir(runtimeRootPath)
	require.NoError(t, dirChangeErr)

	// Ensure stdio example binary is built, as Config validation checks for its existence.
	stdioBinaryPath := filepath.Join(runtimeRootPath, "examples", "demo", "stdio", "my-tool-bin")
	if _, binaryStatErr := os.Stat(stdioBinaryPath); os.IsNotExist(binaryStatErr) {
		t.Logf("Building missing stdio example binary: %s", stdioBinaryPath)
		buildCmdObj := exec.Command("go", "build", "-o", stdioBinaryPath, filepath.Join(runtimeRootPath, "examples", "demo", "stdio", "my-tool", "main.go"))
		buildCmdObj.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "gocache"), "GO111MODULE=off")
		buildCmdObj.Dir = runtimeRootPath
		buildCmdObj.Stdout = os.Stdout
		buildCmdObj.Stderr = os.Stderr
		if buildExecErr := buildCmdObj.Run(); buildExecErr != nil {
			t.Logf("Failed to build stdio example binary (continuing, but validation might fail): %v", buildExecErr)
		}
	}

	// Walk through examples directory.
	walkingErr := filepath.Walk(examplesDirectory, func(exampleConfigPath string, exampleFileInfo os.FileInfo, walkFunctionErr error) error {
		if walkFunctionErr != nil {
			return walkFunctionErr
		}

		if !exampleFileInfo.IsDir() && filepath.Base(exampleConfigPath) == "config.yaml" {
			currentTestCaseName := exampleConfigPath
			if strings.HasPrefix(exampleConfigPath, projectRootPath) {
				currentTestCaseName = strings.TrimPrefix(exampleConfigPath, projectRootPath)
			}
			t.Run(currentTestCaseName, func(subTest *testing.T) {
				validateConfig(subTest, exampleConfigPath)
			})
		}
		return nil
	})
	require.NoError(t, walkingErr)
}

func sourceProjectRoot() (string, error) {
	testWorkspaceName := os.Getenv("TEST_WORKSPACE")
	if testWorkspaceName == "" {
		testWorkspaceName = "_main"
	}
	for _, searchBasePath := range []string{os.Getenv("TEST_SRCDIR"), os.Getenv("RUNFILES_DIR")} {
		if searchBasePath == "" {
			continue
		}
		candidateRootPath := filepath.Join(searchBasePath, testWorkspaceName, "server")
		if _, statCandidateErr := os.Stat(filepath.Join(candidateRootPath, "examples")); statCandidateErr == nil {
			return candidateRootPath, nil
		}
	}
	_, currentFilePath, _, callerOk := runtime.Caller(0)
	if !callerOk {
		return "", os.ErrNotExist
	}
	return filepath.Abs(filepath.Clean(filepath.Join(filepath.Dir(currentFilePath), "../../..")))
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(srcItemPath string, srcItemInfo os.FileInfo, copyWalkFuncErr error) error {
		if copyWalkFuncErr != nil {
			return copyWalkFuncErr
		}
		relativeItemPath, relCalculationErr := filepath.Rel(src, srcItemPath)
		if relCalculationErr != nil {
			return relCalculationErr
		}
		destinationItemPath := filepath.Join(dst, relativeItemPath)
		if srcItemInfo.IsDir() {
			return os.MkdirAll(destinationItemPath, srcItemInfo.Mode())
		}
		inputFileReader, openInputErr := os.Open(srcItemPath)
		if openInputErr != nil {
			return openInputErr
		}
		defer inputFileReader.Close()
		outputFileWriter, createOutputErr := os.OpenFile(destinationItemPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcItemInfo.Mode())
		if createOutputErr != nil {
			return createOutputErr
		}
		defer outputFileWriter.Close()
		_, copyBufferErr := io.Copy(outputFileWriter, inputFileReader)
		return copyBufferErr
	})
}

func validateConfig(targetT *testing.T, targetConfigPath string) {
	aferoFs := afero.NewOsFs()

	requiredEnvironmentVariables := []string{
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

	for _, environmentVariableName := range requiredEnvironmentVariables {
		targetT.Setenv(environmentVariableName, "dummy-val")
	}

	configurationStore := config.NewFileStore(aferoFs, []string{targetConfigPath})
	loadedServiceConfigs, serviceLoadErr := config.LoadServices(context.Background(), configurationStore, "server")
	if serviceLoadErr != nil {
		targetT.Fatalf("Failed to load config %s: %v", targetConfigPath, serviceLoadErr)
	}

	serviceValidationErrors := config.Validate(context.Background(), loadedServiceConfigs, config.Server)
	assert.Empty(targetT, serviceValidationErrors, "Config validation failed for %s", targetConfigPath)
}
