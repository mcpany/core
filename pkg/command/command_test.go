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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalExecutor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		executor := NewExecutor(nil)
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello\n", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("CommandNotFound", func(t *testing.T) {
		executor := NewExecutor(nil)
		_, _, _, err := executor.Execute(context.Background(), "non-existent-command", nil, "", nil)
		assert.Error(t, err)
	})

	t.Run("NonZeroExitCode", func(t *testing.T) {
		executor := NewExecutor(nil)
		_, _, exitCodeChan, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, 1, exitCode)
	})
}

func TestDockerExecutor(t *testing.T) {
	t.Run("WithoutVolumeMount", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello\n", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("WithVolumeMount", func(t *testing.T) {
		// Create a dummy file to mount
		tmpfile, err := os.CreateTemp("", "test-volume-mount")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString("hello from the host")
		require.NoError(t, err)
		tmpfile.Close()

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		containerEnv.SetVolumes(map[string]string{
			"/mnt/test": tmpfile.Name(),
		})
		executor := NewExecutor(containerEnv)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		stdout, stderr, exitCodeChan, err := executor.Execute(ctx, "cat", []string{"/mnt/test"}, "", nil)
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello from the host", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("ImageNotFound", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("non-existent-image:latest")
		executor := NewExecutor(containerEnv)
		_, _, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		assert.Error(t, err)
	})

	t.Run("CommandFailsInContainer", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		_, _, exitCodeChan, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, 1, exitCode)
	})
}

func TestCombinedOutput(t *testing.T) {
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := NewExecutor(containerEnv)
	stdout, stderr, _, err := executor.Execute(context.Background(), "sh", []string{"-c", "echo 'hello stdout' && echo 'hello stderr' >&2"}, "", nil)
	require.NoError(t, err)

	var combined strings.Builder
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		io.Copy(io.MultiWriter(&combined, w), stdout)
	}()
	go func() {
		io.Copy(io.MultiWriter(&combined, w), stderr)
	}()

	_, err = io.ReadAll(r)
	require.NoError(t, err)

	output := combined.String()
	assert.Contains(t, output, "hello stdout")
	assert.Contains(t, output, "hello stderr")
}
