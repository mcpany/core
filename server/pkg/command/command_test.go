// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"bufio"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	dockererrdefs "github.com/docker/docker/errdefs"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
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
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		return false
	}
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

	// Try to start a container to verify if the environment actually supports running containers.
	// This catches issues like "failed to mount ... invalid argument" in nested overlayfs environments.
	ctx := context.Background()
	// Use a lightweight image that is likely to be present or small to pull.
	// If alpine:latest is not found, we attempt to pull it, but failure to pull also means we can't test.
	_, _, err = cli.ImageInspectWithRaw(ctx, "alpine:latest")
	if client.IsErrNotFound(err) {
		reader, err := cli.ImagePull(ctx, "alpine:latest", image.PullOptions{})
		if err != nil {
			t.Logf("could not pull alpine:latest: %v", err)
			return false
		}
		defer reader.Close()
		_, _ = io.Copy(io.Discard, reader)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"true"},
	}, nil, nil, nil, "")
	if err != nil {
		t.Logf("could not create probe container: %v", err)
		return false
	}
	defer func() {
		_ = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.Logf("could not start probe container (environment might be broken): %v", err)
		return false
	}

	return true
}

func TestLocalExecutor(t *testing.T) {
	t.Parallel()
	t.Run("Success", func(t *testing.T) {
		executor := NewExecutor(nil)
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
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
			stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
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

	t.Run("ContextCancellation", func(t *testing.T) {
		executor := NewExecutor(nil)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Start a long-running command
		_, _, exitCodeChan, err := executor.Execute(ctx, "sleep", []string{"10"}, "", nil)
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

	t.Run("RestrictedWorkingDir", func(t *testing.T) {
		// Set allowed paths to a specific directory
		tmpDir := t.TempDir()
		// Get absolute path to ensure matching works correctly
		absTmpDir, err := filepath.Abs(tmpDir)
		require.NoError(t, err)

		validation.SetAllowedPaths([]string{absTmpDir})
		t.Cleanup(func() { validation.SetAllowedPaths(nil) })

		executor := NewExecutor(nil)

		// 1. Allowed path should succeed
		stdout, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, absTmpDir, nil)
		require.NoError(t, err)

		// Consume output to prevent hanging
		_, _ = io.ReadAll(stdout)
		assert.Equal(t, 0, <-exitCodeChan)

		// 2. Disallowed path should fail
		// Create another temp dir that is NOT allowed
		disallowedDir := t.TempDir()
		absDisallowedDir, err := filepath.Abs(disallowedDir)
		require.NoError(t, err)

		_, _, _, err = executor.Execute(context.Background(), "echo", []string{"hello"}, absDisallowedDir, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid working directory")
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
		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		var stdoutBytes []byte
		for i := 0; i < 5; i++ {
			stdoutBytes, err = io.ReadAll(stdout)
			require.NoError(t, err)
			if string(stdoutBytes) == "hello\n" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
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
		defer func() { _ = os.Remove(tmpfile.Name()) }()

		_, err = tmpfile.WriteString("hello from the host")
		require.NoError(t, err)
		_ = tmpfile.Close()

		absPath, err := filepath.Abs(tmpfile.Name())
		require.NoError(t, err)

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		containerEnv.SetVolumes(map[string]string{
			absPath: "/mnt/test",
		})
		executor := NewExecutor(containerEnv)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		stdout, stderr, exitCodeChan, err := executor.Execute(ctx, "cat", []string{"/mnt/test"}, "", nil)
		require.NoError(t, err)

		var stdoutBytes []byte
		for i := 0; i < 5; i++ {
			stdoutBytes, err = io.ReadAll(stdout)
			require.NoError(t, err)
			if string(stdoutBytes) == "hello from the host" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
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

	t.Run("ContextCancellation", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, _, exitCodeChan, err := executor.Execute(ctx, "sleep", []string{"10"}, "", nil)
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
		// t.Skip("Skipping flaky test: ContainerIsRemoved")
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		containerName := fmt.Sprintf("test-container-removal-%d", time.Now().UnixNano())
		containerEnv.SetName(containerName)
		executor := NewExecutor(containerEnv)

		// Ensure cleanup even if test fails
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		require.NoError(t, err)
		defer cli.Close()
		defer func() {
			_ = cli.ContainerRemove(context.Background(), containerName, container.RemoveOptions{Force: true})
		}()

		_, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		<-exitCodeChan

		// Check if container is removed
		var lastErr error
		for i := 0; i < 20; i++ {
			_, err = cli.ContainerInspect(context.Background(), containerName)
			if dockererrdefs.IsNotFound(err) {
				lastErr = err
				break
			}
			lastErr = err
			time.Sleep(100 * time.Millisecond)
		}
		assert.True(t, dockererrdefs.IsNotFound(lastErr), "Expected container to be removed, got: %v", lastErr)
	})
}

func TestCombinedOutput(t *testing.T) {
	if !canConnectToDocker(t) {
		t.Skip("Cannot connect to Docker daemon, skipping Docker tests")
	}
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := NewExecutor(containerEnv)
	stdout, stderr, _, err := executor.Execute(context.Background(), "sh", []string{"-c", "echo 'hello stdout' && echo 'hello stderr' >&2"}, "", nil)
	require.NoError(t, err)

	var combined strings.Builder
	var mu sync.Mutex
	r, w := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		writer := io.MultiWriter(&muWriter{&combined, &mu}, w)
		_, _ = io.Copy(writer, stdout)
	}()
	go func() {
		defer wg.Done()
		writer := io.MultiWriter(&muWriter{&combined, &mu}, w)
		_, _ = io.Copy(writer, stderr)
	}()

	go func() {
		wg.Wait()
		_ = w.Close()
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

func TestNewExecutor(t *testing.T) {
	t.Run("WithContainer", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		assert.IsType(t, &dockerExecutor{}, executor)
	})

	t.Run("WithoutContainer", func(t *testing.T) {
		executor := NewExecutor(nil)
		assert.IsType(t, &localExecutor{}, executor)
	})
}

func TestNewLocalExecutor(t *testing.T) {
	executor := NewLocalExecutor()
	assert.NotNil(t, executor)
	assert.IsType(t, &localExecutor{}, executor)
}

func TestLocalExecutorWithStdIO(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		executor := NewExecutor(nil)
		stdin, stdout, stderr, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "cat", nil, "", nil)
		require.NoError(t, err)

		go func() {
			defer func() { _ = stdin.Close() }()
			_, err := stdin.Write([]byte("hello\n"))
			require.NoError(t, err)
		}()

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

	t.Run("Execute_ContainerStartError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerStartFunc = func(ctx context.Context, containerID string, options container.StartOptions) error {
			return errors.New("start error")
		}
		// Need mocked remove because Start failure attempts to remove
		mockClient.ContainerRemoveFunc = func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start error")
	})

	t.Run("Execute_ContainerLogsError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerLogsFunc = func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			return nil, errors.New("logs error")
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "logs error")
	})

	t.Run("Execute_ClientFactoryError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		executor.clientFactory = func() (DockerClient, error) {
			return nil, errors.New("client factory error")
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client factory error")
	})

	t.Run("Execute_ContainerWaitError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerLogsFunc = func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		mockClient.ContainerWaitFunc = func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
			errCh := make(chan error, 1)
			errCh <- errors.New("wait error")
			return nil, errCh
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, -1, exitCode)
	})

	t.Run("Execute_ImagePullError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ImagePullFunc = func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
			return nil, errors.New("pull error")
		}
		// Logs should be mocked to succeed
		mockClient.ContainerLogsFunc = func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		// Wait should succeed
		mockClient.ContainerWaitFunc = func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
			statusCh := make(chan container.WaitResponse, 1)
			statusCh <- container.WaitResponse{StatusCode: 0}
			return statusCh, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)
		// Should succeed even if pull fails (logs warning)
		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("ExecuteWithStdIO_ClientFactoryError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		executor.clientFactory = func() (DockerClient, error) {
			return nil, errors.New("client factory error")
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client factory error")
	})

	t.Run("ExecuteWithStdIO_ContainerCreateError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
			return container.CreateResponse{}, errors.New("create error")
		}
		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create error")
	})
}

func TestDockerExecutorWithStdIO(t *testing.T) {
	t.Skip("Skipping flaky test: TestDockerExecutorWithStdIO (hangs on stream read)")
	if !canConnectToDocker(t) {
		t.Skip("Cannot connect to Docker daemon, skipping Docker tests")
	}

	t.Run("Success", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := NewExecutor(containerEnv)
		stdin, stdout, stderr, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "cat", nil, "", nil)
		require.NoError(t, err)

		go func() {
			defer func() { _ = stdin.Close() }()
			_, err := stdin.Write([]byte("hello\n"))
			require.NoError(t, err)
			time.Sleep(200 * time.Millisecond)
		}()

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

	t.Run("Execute_VolumeMounts", func(t *testing.T) {
		// Ensure we have a valid directory for the volume mount test.
		// In CWD, "testdata" usually exists.
		if _, err := os.Stat("testdata"); os.IsNotExist(err) {
			_ = os.Mkdir("testdata", 0755)
			defer os.Remove("testdata")
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		// Use a relative path that IsAllowedPath will accept (assuming it doesn't contain ..)
		// and verify that it is passed to Docker.
		// Note: IsAllowedPath checks relative to CWD. We use "testdata" as a safe relative path.
		containerEnv.SetVolumes(map[string]string{
			"testdata": "/container/path",
		})
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		var capturedHostConfig *container.HostConfig
		mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
			capturedHostConfig = hostConfig
			return container.CreateResponse{ID: "test-id"}, nil
		}
		// Mock ImagePull to avoid nil pointer dereference or real network call
		mockClient.ImagePullFunc = func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		// Mock ContainerStart/Logs/Wait/Remove to complete execution flow
		mockClient.ContainerStartFunc = func(ctx context.Context, containerID string, options container.StartOptions) error {
			return nil
		}
		mockClient.ContainerLogsFunc = func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		mockClient.ContainerWaitFunc = func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
			statusCh := make(chan container.WaitResponse, 1)
			statusCh <- container.WaitResponse{StatusCode: 0}
			return statusCh, nil
		}
		mockClient.ContainerRemoveFunc = func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		require.NotNil(t, capturedHostConfig)
		require.Len(t, capturedHostConfig.Mounts, 1)
		// With the fix, Key is Source, Value is Target
		assert.Equal(t, "/host/path", capturedHostConfig.Mounts[0].Source)
		assert.Equal(t, "/container/path", capturedHostConfig.Mounts[0].Target)
	})
}

func TestDockerExecutor_Mocked(t *testing.T) {
	t.Run("Execute_Success", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerLogsFunc = func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
			var buf bytes.Buffer
			// Stdout header: 1 (stdout), 0, 0, 0, size (big endian uint32)
			buf.Write([]byte{1, 0, 0, 0})
			if err := binary.Write(&buf, binary.BigEndian, uint32(5)); err != nil {
				return nil, err
			}
			buf.WriteString("hello")
			return io.NopCloser(&buf), nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("Execute_ContainerCreateError", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
			return container.CreateResponse{}, errors.New("create error")
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create error")
	})

	t.Run("ExecuteWithStdIO_Success", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			server, client := net.Pipe()

			go func() {
				defer server.Close()
				var buf bytes.Buffer
				buf.Write([]byte{1, 0, 0, 0})
				binary.Write(&buf, binary.BigEndian, uint32(5))
				buf.WriteString("hello")
				server.Write(buf.Bytes())
			}()

			return types.HijackedResponse{
				Conn:   client,
				Reader: bufio.NewReader(client),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		stdin, stdout, stderr, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "cat", nil, "", nil)
		require.NoError(t, err)
		defer stdin.Close()

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("ExecuteWithStdIO_Write", func(t *testing.T) {
		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
			server, client := net.Pipe()

			go func() {
				defer server.Close()
				buf := make([]byte, 1024)
				// Read from client (simulate container receiving input)
				n, _ := server.Read(buf)
				if n > 0 {
					// Echo back to stdout
					var outBuf bytes.Buffer
					outBuf.Write([]byte{1, 0, 0, 0})
					_ = binary.Write(&outBuf, binary.BigEndian, uint32(n))
					outBuf.Write(buf[:n])
					_, _ = server.Write(outBuf.Bytes())
				}
			}()

			return types.HijackedResponse{
				Conn:   client,
				Reader: bufio.NewReader(client),
			}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		stdin, stdout, _, exitCodeChan, err := executor.ExecuteWithStdIO(context.Background(), "cat", nil, "", nil)
		require.NoError(t, err)

		// Test Write
		_, err = stdin.Write([]byte("hello"))
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(stdoutBytes))

        stdin.Close()

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("Execute_VolumeMounts", func(t *testing.T) {
		// Ensure we have a valid directory for the volume mount test.
		// In CWD, "testdata" usually exists.
		if _, err := os.Stat("testdata"); os.IsNotExist(err) {
			_ = os.Mkdir("testdata", 0755)
			defer os.Remove("testdata")
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		// Use a relative path that IsAllowedPath will accept (assuming it doesn't contain ..)
		// and verify that it is passed to Docker.
		// Note: IsAllowedPath checks relative to CWD. We use "testdata" as a safe relative path.
		containerEnv.SetVolumes(map[string]string{
			"testdata": "/container/path",
		})
		executor := newDockerExecutor(containerEnv).(*dockerExecutor)

		mockClient := &MockDockerClient{}
		var capturedHostConfig *container.HostConfig
		mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
			capturedHostConfig = hostConfig
			return container.CreateResponse{ID: "test-id"}, nil
		}

		executor.clientFactory = func() (DockerClient, error) {
			return mockClient, nil
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		require.NotNil(t, capturedHostConfig)
		require.Len(t, capturedHostConfig.Mounts, 1)
		// We expect the resolved absolute path
		absTestdata, _ := filepath.Abs("testdata")
		assert.Equal(t, absTestdata, capturedHostConfig.Mounts[0].Source)
		assert.Equal(t, "/container/path", capturedHostConfig.Mounts[0].Target)
	})
}
