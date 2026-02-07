// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTempConfigFile(t *testing.T) {
	t.Parallel()
	// Create a dummy config
	serviceName := "test-service"
	config := &configv1.UpstreamServiceConfig{}
	config.SetName(serviceName)

	// Create a dummy config file
	configFilePath := CreateTempConfigFile(t, config)
	defer func() { _ = os.Remove(configFilePath) }()

	// Verify that the file was created
	_, err := os.Stat(configFilePath)
	require.NoError(t, err, "Temp config file should exist")
}

func TestProjectRoot(t *testing.T) {
	t.Parallel()
	// Get the project root
	root := ProjectRoot(t)

	// Verify that the project root contains a known file
	_, err := os.Stat(filepath.Join(root, "go.mod"))
	require.NoError(t, err, "go.mod should exist in the project root")
}

func TestManagedProcess(t *testing.T) {
	t.Parallel()
	mp := NewManagedProcess(t, "echo", "echo", []string{"hello"}, nil)
	require.NoError(t, mp.Start())
	// Wait for the process to finish on its own
	<-mp.waitDone
	assert.Contains(t, mp.StdoutString(), "hello")
	// Calling stop on an already finished process should be a no-op
	mp.Stop()
}

func TestWaitForText(t *testing.T) {
	t.Parallel()
	mp := NewManagedProcess(t, "echo", "sh", []string{"-c", "sleep 0.1; echo hello"}, nil)
	require.NoError(t, mp.Start())
	mp.WaitForText(t, "hello", 1*time.Second)
	mp.Stop()
}

func TestDockerHelpers(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping Docker tests because SKIP_DOCKER_TESTS is set")
	}
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Skipping TestDockerHelpers in CI due to potential rate limiting/network issues")
	}
	t.Parallel()
	if !IsDockerSocketAccessible() {
		t.Skip("Docker is not available")
	}

	// Test StartDockerContainer
	imageName := "alpine:latest"
	containerName := fmt.Sprintf("mcpany-test-container-%d", time.Now().UnixNano())
	cleanup := StartDockerContainer(t, imageName, containerName, []string{"-d"}, "sleep", "60")
	defer cleanup()

	// Verify the container is running
	dockerExe, dockerArgs := getDockerCommand()
	psCmd := exec.Command(dockerExe, append(dockerArgs, "ps", "-f", fmt.Sprintf("name=%s", containerName))...) //nolint:gosec // Test helper
	out, err := psCmd.Output()
	require.NoError(t, err, "docker ps command failed. Output: %s", string(out))
	assert.Contains(t, string(out), containerName)

	// Test StartRedisContainer
	_, redisCleanup := StartRedisContainer(t)
	defer redisCleanup()

	// The StartRedisContainer function has internal checks to ensure the container
	// starts and is responsive. A successful return is a pass.
}
