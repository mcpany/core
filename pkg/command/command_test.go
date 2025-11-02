/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerExecutor_VolumeMount(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("Skipping test that requires Docker daemon")
	}
	// Create a temporary file with some content.
	tmpfile, err := ioutil.TempFile("", "test-volume-mount-")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := "hello from the test file"
	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	// Create a docker executor with a volume mount.
	image := "alpine:latest"
	containerEnv := configv1.ContainerEnvironment_builder{
		Image: &image,
		Volumes: map[string]string{
			"/test-file": tmpfile.Name(),
		},
	}.Build()
	executor := NewExecutor(containerEnv)

	// Execute the "cat" command in the container to read the mounted file.
	result, _, err := executor.Execute(context.Background(), "cat", []string{"/test-file"}, "", nil)
	require.NoError(t, err)
	defer result.Combined.Close()

	// Read the output from the command.
	output, err := ioutil.ReadAll(result.Combined)
	require.NoError(t, err)

	// Check that the output from the command is the same as the content of the file.
	assert.Equal(t, content, string(output))
}

func TestLocalExecutor_Execute(t *testing.T) {
	executor := NewExecutor(nil)
	result, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
	require.NoError(t, err)
	defer result.Combined.Close()

	output, err := ioutil.ReadAll(result.Combined)
	require.NoError(t, err)

	assert.Equal(t, "hello\n", string(output))
}

func TestNewExecutor_Docker(t *testing.T) {
	image := "alpine:latest"
	containerEnv := configv1.ContainerEnvironment_builder{
		Image: &image,
	}.Build()
	executor := NewExecutor(containerEnv)
	assert.IsType(t, &dockerExecutor{}, executor)
}

func TestNewExecutor_Local(t *testing.T) {
	executor := NewExecutor(nil)
	assert.IsType(t, &localExecutor{}, executor)
}

func TestLocalExecutor_Execute_Error(t *testing.T) {
	executor := NewExecutor(nil)
	_, exitCode, err := executor.Execute(context.Background(), "command-that-does-not-exist", nil, "", nil)
	assert.Error(t, err)
	assert.NotEqual(t, 0, exitCode)
}

func TestLocalExecutor_Execute_ExitCode(t *testing.T) {
	executor := NewExecutor(nil)
	_, exitCode, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil)
	assert.Error(t, err)
	assert.Equal(t, 1, exitCode)
}

func TestDockerExecutor_Execute_Error(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("Skipping test that requires Docker daemon")
	}

	image := "command-that-does-not-exist"
	containerEnv := configv1.ContainerEnvironment_builder{
		Image: &image,
	}.Build()
	executor := NewExecutor(containerEnv)

	_, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
	assert.Error(t, err)
}

func TestDockerExecutor_Execute_StartError(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("Skipping test that requires Docker daemon")
	}

	image := "alpine:latest"
	containerEnv := configv1.ContainerEnvironment_builder{
		Image: &image,
	}.Build()
	executor := NewExecutor(containerEnv)

	_, exitCode, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil)
	assert.Error(t, err)
	assert.Equal(t, 1, exitCode)
}

func TestDockerExecutor_Execute_ContextCancel(t *testing.T) {
	if os.Getenv("DOCKER_HOST") == "" {
		t.Skip("Skipping test that requires Docker daemon")
	}

	image := "alpine:latest"
	containerEnv := configv1.ContainerEnvironment_builder{
		Image: &image,
	}.Build()
	executor := NewExecutor(containerEnv)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := executor.Execute(ctx, "sleep", []string{"1"}, "", nil)
	assert.Error(t, err)
}
