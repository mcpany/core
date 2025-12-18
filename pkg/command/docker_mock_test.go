// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockDockerClient struct {
	ImagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStartFunc  func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRemoveFunc func(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogsFunc   func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerWaitFunc   func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerAttachFunc func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	CloseFunc           func() error
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	if m.ImagePullFunc != nil {
		return m.ImagePullFunc(ctx, ref, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if m.ContainerCreateFunc != nil {
		return m.ContainerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{ID: "test-container-id"}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	if m.ContainerStartFunc != nil {
		return m.ContainerStartFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.ContainerRemoveFunc != nil {
		return m.ContainerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	if m.ContainerLogsFunc != nil {
		return m.ContainerLogsFunc(ctx, container, options)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	if m.ContainerWaitFunc != nil {
		return m.ContainerWaitFunc(ctx, containerID, condition)
	}
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	statusCh <- container.WaitResponse{StatusCode: 0}
	close(statusCh)
	close(errCh)
	return statusCh, errCh
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	if m.ContainerAttachFunc != nil {
		return m.ContainerAttachFunc(ctx, container, options)
	}
	return types.HijackedResponse{
		Conn:   &mockConn{},
		Reader: nil,
	}, nil
}

func (m *MockDockerClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

type mockConn struct {
	net.Conn
}

func (m *mockConn) Write(b []byte) (n int, err error) { return len(b), nil }
func (m *mockConn) Close() error                      { return nil }
func (m *mockConn) CloseWrite() error                 { return nil }

func TestDockerExecutor_Mock(t *testing.T) {
	t.Run("Execute_Success", func(t *testing.T) {
		mockClient := &MockDockerClient{
			ContainerLogsFunc: func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
				// Construct a stdcopy frame for stdout
				content := "mock output"
				header := make([]byte, 8)
				header[0] = 1 // stdout
				binary.BigEndian.PutUint32(header[4:], uint32(len(content)))
				return io.NopCloser(io.MultiReader(bytes.NewReader(header), strings.NewReader(content))), nil
			},
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := &dockerExecutor{
			containerEnv: containerEnv,
			clientFactory: func() (DockerClient, error) {
				return mockClient, nil
			},
		}

		stdout, stderr, exitCodeChan, err := executor.Execute(context.Background(), "echo", []string{"hello"}, "", nil)
		require.NoError(t, err)

		stdoutBytes, err := io.ReadAll(stdout)
		require.NoError(t, err)
		assert.Equal(t, "mock output", string(stdoutBytes))

		stderrBytes, err := io.ReadAll(stderr)
		require.NoError(t, err)
		assert.Empty(t, string(stderrBytes))

		exitCode := <-exitCodeChan
		assert.Equal(t, 0, exitCode)
	})

	t.Run("Execute_CreateFail", func(t *testing.T) {
		mockClient := &MockDockerClient{
			ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{}, fmt.Errorf("create failed")
			},
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := &dockerExecutor{
			containerEnv: containerEnv,
			clientFactory: func() (DockerClient, error) {
				return mockClient, nil
			},
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.ErrorContains(t, err, "create failed")
	})

	t.Run("Execute_StartFail", func(t *testing.T) {
		mockClient := &MockDockerClient{
			ContainerStartFunc: func(ctx context.Context, containerID string, options container.StartOptions) error {
				return fmt.Errorf("start failed")
			},
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := &dockerExecutor{
			containerEnv: containerEnv,
			clientFactory: func() (DockerClient, error) {
				return mockClient, nil
			},
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.ErrorContains(t, err, "start failed")
	})

	t.Run("Execute_LogsFail", func(t *testing.T) {
		mockClient := &MockDockerClient{
			ContainerLogsFunc: func(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
				return nil, fmt.Errorf("logs failed")
			},
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := &dockerExecutor{
			containerEnv: containerEnv,
			clientFactory: func() (DockerClient, error) {
				return mockClient, nil
			},
		}

		_, _, _, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		assert.ErrorContains(t, err, "logs failed")
	})

	t.Run("Execute_WaitFail", func(t *testing.T) {
		mockClient := &MockDockerClient{
			ContainerWaitFunc: func(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
				errCh := make(chan error, 1)
				errCh <- fmt.Errorf("wait failed")
				close(errCh)
				statusCh := make(chan container.WaitResponse)
				// Do not close statusCh to force selection of errCh
				return statusCh, errCh
			},
		}

		containerEnv := &configv1.ContainerEnvironment{}
		containerEnv.SetImage("alpine:latest")
		executor := &dockerExecutor{
			containerEnv: containerEnv,
			clientFactory: func() (DockerClient, error) {
				return mockClient, nil
			},
		}

		_, _, exitCodeChan, err := executor.Execute(context.Background(), "echo", nil, "", nil)
		require.NoError(t, err)

		exitCode := <-exitCodeChan
		assert.Equal(t, -1, exitCode)
	})
}
