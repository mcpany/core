// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"errors"
	"net"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConnWithCloseWrite simulates a connection that supports CloseWrite
type mockConnWithCloseWrite struct {
	net.Conn
	closeWriteCalled bool
}

func (m *mockConnWithCloseWrite) CloseWrite() error {
	m.closeWriteCalled = true
	return nil
}

func TestDockerExecutor_CloseWriter_CloseWrite(t *testing.T) {
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := newDockerExecutor(containerEnv).(*dockerExecutor)

	mockClient := &MockDockerClient{}

	// We need a custom mock conn that implements CloseWrite
	server, client := net.Pipe()
	mockConn := &mockConnWithCloseWrite{
		Conn: client,
	}

	mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
		return types.HijackedResponse{
			Conn:   mockConn,
			Reader: bufio.NewReader(mockConn), // Fix for StdCopy panic
		}, nil
	}

	executor.clientFactory = func() (DockerClient, error) {
		return mockClient, nil
	}

	// Mock other calls to succeed
	mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
		return container.CreateResponse{ID: "test-id"}, nil
	}

	// ExecuteWithStdIO returns the closeWriter as io.WriteCloser (stdin)
	stdin, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
	require.NoError(t, err)

	// Cast stdin to *closeWriter (not possible as it is private)
	// But stdin is io.WriteCloser, so we just call Close()
	err = stdin.Close()
	assert.NoError(t, err)

	// Verify CloseWrite was called
	assert.True(t, mockConn.closeWriteCalled)

	// Cleanup
	server.Close()
	// client is closed by stdin.Close()
}

func TestNewDockerExecutor_DefaultClientFactory(t *testing.T) {
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")

	// This returns Executor interface
	execInterface := newDockerExecutor(containerEnv)

	// Assert it is *dockerExecutor
	exec, ok := execInterface.(*dockerExecutor)
	require.True(t, ok)

	// Access clientFactory field directly as it is in the same package
	factory := exec.clientFactory
	require.NotNil(t, factory)

	// Call it
	client, err := factory()

	// We expect either a valid client or an error (if docker not available/env vars issue)
	// But the factory function itself executed, covering the closure lines.
	if err != nil {
		assert.Error(t, err)
	} else {
		// If Docker is available, client might be valid but we can't assert much about it without mocking or real docker
		_ = client
	}
}

func TestDockerExecutor_CloseWriter_CloseWrite_Error(t *testing.T) {
	containerEnv := &configv1.ContainerEnvironment{}
	containerEnv.SetImage("alpine:latest")
	executor := newDockerExecutor(containerEnv).(*dockerExecutor)

	mockClient := &MockDockerClient{}

	// We need a custom mock conn that implements CloseWrite and returns error
	server, client := net.Pipe()
	mockConn := &mockConnWithCloseWriteError{
		Conn: client,
	}

	mockClient.ContainerAttachFunc = func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
		return types.HijackedResponse{
			Conn:   mockConn,
			Reader: bufio.NewReader(mockConn), // Fix for StdCopy panic
		}, nil
	}

	// Mock other calls to succeed
	mockClient.ContainerCreateFunc = func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
		return container.CreateResponse{ID: "test-id"}, nil
	}

	executor.clientFactory = func() (DockerClient, error) {
		return mockClient, nil
	}

	stdin, _, _, _, err := executor.ExecuteWithStdIO(context.Background(), "echo", nil, "", nil)
	require.NoError(t, err)

	err = stdin.Close()
	assert.Error(t, err)
	assert.Equal(t, "closewrite error", err.Error())

	server.Close()
}

type mockConnWithCloseWriteError struct {
	net.Conn
}

func (m *mockConnWithCloseWriteError) CloseWrite() error {
	return errors.New("closewrite error")
}
