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
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type muWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func (w *muWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

func canConnectToDocker(t *testing.T) bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Logf("could not create docker client: %v", err)
		return false
	}
	_, err = cli.Ping(context.Background())
	if err != nil {
		t.Logf("could not ping docker daemon: %v", err)
		return false
	}
	return true
}

func TestLocalExecutor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		executor := NewExecutor(nil)
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil, nil)
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)
		var stdoutBytes, stderrBytes []byte
		var stdoutErr, stderrErr error

		go func() {
			defer wg.Done()
			stdoutBytes, stdoutErr = io.ReadAll(stdout)
		}()

		go func() {
			defer wg.Done()
			stderrBytes, stderrErr = io.ReadAll(stderr)
		}()

		wg.Wait()

		require.NoError(t, stdoutErr)
		assert.Equal(t, "hello\n", string(stdoutBytes))

		require.NoError(t, stderrErr)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("GoroutineExecution", func(t *testing.T) {
		t.Parallel()
		for i := 0; i < 10; i++ {
			executor := NewExecutor(nil)
			stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil, nil)
			require.NoError(t, err)

			var wg sync.WaitGroup
			wg.Add(2)
			var stdoutBytes, stderrBytes []byte
			var stdoutErr, stderrErr error

			go func() {
				defer wg.Done()
				stdoutBytes, stdoutErr = io.ReadAll(stdout)
			}()

			go func() {
				defer wg.Done()
				stderrBytes, stderrErr = io.ReadAll(stderr)
			}()

			wg.Wait()

			require.NoError(t, stdoutErr)
			assert.Equal(t, "hello\n", string(stdoutBytes))

			require.NoError(t, stderrErr)
			assert.Empty(t, string(stderrBytes))

			exitCode := <-exitCodeChan
			assert.Equal(t, 0, exitCode)
		}
	})

	t.Run("CommandNotFound", func(t *testing.T) {
		executor := NewExecutor(nil)
		_, _, _, err := executor.Execute(context.Background(), "non-existent-command", nil, "", nil, nil)
		assert.Error(t, err)
	})

	t.Run("NonZeroExitCode", func(t *testing.T) {
		executor := NewExecutor(nil)
		_, _, exitCodeChan, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil, nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, 1, exitCode)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		executor := NewExecutor(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start a long-running command
		_, _, exitCodeChan, err := executor.Execute(ctx, "sleep", []string{"10"}, "", nil, nil)
		require.NoError(t, err)

		// Cancel the context almost immediately
		cancel()

		// Expect the command to be terminated and receive a non-zero exit code
		select {
		case exitCode := <-exitCodeChan:
			assert.NotEqual(t, 0, exitCode, "Expected a non-zero exit code due to context cancellation")
		case <-time.After(2 * time.Second):
			t.Fatal("Test timed out waiting for command to exit")
		}
	})
}

func TestDockerExecutor(t *testing.T) {
	if !canConnectToDocker(t) {
		t.Skip("Cannot connect to Docker daemon, skipping Docker tests")
	}
	t.Run("WithoutVolumeMount", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil, nil)
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
		tmpfile, err := os.CreateTemp(".", "test-volume-mount")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.WriteString("hello from the host")
		require.NoError(t, err)
		tmpfile.Close()

		absPath, err := filepath.Abs(tmpfile.Name())
		require.NoError(t, err)

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		containerEnv.SetVolumes(map[string]string{
			"/mnt/test": absPath,
		})
		executor := NewExecutor(containerEnv)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		stdout, stderr, exitCodeChan, err := executor.Execute(ctx, "cat", []string{"/mnt/test"}, "", nil, nil)
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
		_, _, _, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil, nil)
		assert.Error(t, err)
	})

	t.Run("CommandFailsInContainer", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		_, _, exitCodeChan, err := executor.Execute(context.Background(), "sh", []string{"-c", "exit 1"}, "", nil, nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, 1, exitCode)
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, _, exitCodeChan, err := executor.Execute(ctx, "sleep", []string{"10"}, "", nil, nil)
		require.NoError(t, err)

		cancel()

		select {
		case exitCode := <-exitCodeChan:
			assert.NotEqual(t, 0, exitCode, "Expected a non-zero exit code due to context cancellation")
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for command to exit")
		}
	})

	t.Run("ContainerIsRemoved", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		containerEnv.SetName("test-container-removal")
		executor := NewExecutor(containerEnv)
		_, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil, nil)
		require.NoError(t, err)

		<-exitCodeChan

		// Check if container is removed
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		require.NoError(t, err)
		_, err = cli.ContainerInspect(context.Background(), "test-container-removal")
		assert.True(t, client.IsErrNotFound(err), "Expected container to be removed")
	})
}

func TestCombinedOutput(t *testing.T) {
	if !canConnectToDocker(t) {
		t.Skip("Cannot connect to Docker daemon, skipping Docker tests")
	}
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := NewExecutor(containerEnv)
	stdout, stderr, _, err := executor.Execute(context.Background(), "sh", []string{"-c", "echo 'hello stdout' && echo 'hello stderr' >&2"}, "", nil, nil)
	require.NoError(t, err)

	var combined strings.Builder
	var mu sync.Mutex
	r, w := io.Pipe()

	go func() {
		defer w.Close()
		writer := io.MultiWriter(&muWriter{&combined, &mu}, w)
		io.Copy(writer, stdout)
	}()
	go func() {
		writer := io.MultiWriter(&muWriter{&combined, &mu}, w)
		io.Copy(writer, stderr)
	}()

	_, err = io.ReadAll(r)
	require.NoError(t, err)

	output := combined.String()
	assert.Contains(t, output, "hello stdout")
	assert.Contains(t, output, "hello stderr")
}

func TestNewDockerExecutorSuccess(t *testing.T) {
	if !canConnectToDocker(t) {
		t.Skip("Cannot connect to Docker daemon, skipping Docker tests")
	}
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := newDockerExecutor(containerEnv)
	assert.NotNil(t, executor)
}
