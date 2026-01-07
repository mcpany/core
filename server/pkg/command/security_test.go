// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestDockerExecutor_Security_Execute_InvalidVolume(t *testing.T) {
    name := "test-container"
    image := "alpine"
	// Setup with invalid volume path
	containerEnv := &configv1.ContainerEnvironment{
		Name:  &name,
		Image: &image,
		Volumes: map[string]string{
			"../../bad": "/mnt/bad", // Vulnerability: Path traversal
		},
	}

	executor := newDockerExecutor(containerEnv)
	dockerExec, ok := executor.(*dockerExecutor)
	assert.True(t, ok)

	// Inject Mock Client (should not be called if validation works, but needed for setup)
	dockerExec.clientFactory = func() (DockerClient, error) {
		return &MockDockerClient{}, nil
	}

	// Execution
	_, _, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "/", nil)

	// Verification
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "insecure volume mount")
		assert.Contains(t, err.Error(), "../../bad")
	}
}

func TestDockerExecutor_Security_ExecuteWithStdIO_InvalidVolume(t *testing.T) {
    name := "test-container"
    image := "alpine"
	// Setup with invalid volume path (absolute path outside CWD)
	containerEnv := &configv1.ContainerEnvironment{
		Name:  &name,
		Image: &image,
		Volumes: map[string]string{
			"/etc/passwd": "/mnt/passwd", // Vulnerability: Mounting sensitive host file
		},
	}

	executor := newDockerExecutor(containerEnv)
	dockerExec, ok := executor.(*dockerExecutor)
	assert.True(t, ok)

	// Inject Mock Client
	dockerExec.clientFactory = func() (DockerClient, error) {
		return &MockDockerClient{}, nil
	}

	// Execution
	_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", []string{"hello"}, "/", nil)

	// Verification
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "insecure volume mount")
		assert.Contains(t, err.Error(), "/etc/passwd")
	}
}
