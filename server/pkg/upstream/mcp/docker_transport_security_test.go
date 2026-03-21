// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockNetConn struct {
	io.Reader
	io.Writer
}

func (m *mockNetConn) Close() error                       { return nil }
func (m *mockNetConn) LocalAddr() net.Addr                { return nil }
func (m *mockNetConn) RemoteAddr() net.Addr               { return nil }
func (m *mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error { return nil }

func TestDockerTransport_Security_SetupCommands(t *testing.T) {
	// Mock Docker Client
	originalNewDockerClient := newDockerClient
	var capturedCmd []string
	newDockerClient = func(_ ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(_ context.Context, _ string, _ image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(_ context.Context, config *container.Config, _ *container.HostConfig, _ *network.NetworkingConfig, _ *v1.Platform, _ string) (container.CreateResponse, error) {
				capturedCmd = config.Cmd
				return container.CreateResponse{ID: "test-container-id"}, nil
			},
			ContainerAttachFunc: func(_ context.Context, _ string, _ container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{
					Conn:   &mockNetConn{Reader: bytes.NewReader([]byte{}), Writer: io.Discard},
					Reader: bufio.NewReader(&bytes.Buffer{}),
				}, nil
			},
			ContainerStartFunc: func(_ context.Context, _ string, _ container.StartOptions) error {
				return nil
			},
			ContainerStopFunc: func(_ context.Context, _ string, _ container.StopOptions) error {
				return nil
			},
			ContainerRemoveFunc: func(_ context.Context, _ string, _ container.RemoveOptions) error {
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	maliciousCommand := "echo POISONED"
	stdioConfig := configv1.McpStdioConnection_builder{
		ContainerImage: proto.String("alpine:latest"),
		Command:        proto.String("echo"),
		Args:           []string{"hello"},
		SetupCommands:  []string{maliciousCommand},
	}.Build()

	t.Run("Default Secure Behavior", func(t *testing.T) {
		// Ensure flag is unset
		os.Unsetenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS")
		capturedCmd = nil // Reset captured command

		transport := &DockerTransport{StdioConfig: stdioConfig}
		_, err := transport.Connect(context.Background())

		// Assertions
		require.Error(t, err, "Connect should fail when setup_commands are present and flag is unset")
		assert.Contains(t, err.Error(), "setup_commands are disabled by default", "Error should explain security restriction")
		assert.Nil(t, capturedCmd, "Container should NOT have been created")
	})

	t.Run("Explicit Opt-in", func(t *testing.T) {
		// Set flag to true
		os.Setenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS", "true")
		defer os.Unsetenv("MCP_ALLOW_UNSAFE_SETUP_COMMANDS")
		capturedCmd = nil // Reset captured command

		transport := &DockerTransport{StdioConfig: stdioConfig}
		_, err := transport.Connect(context.Background())

		// Assertions
		require.NoError(t, err, "Connect should succeed when flag is set")
		require.NotNil(t, capturedCmd, "Container should have been created")
		require.Len(t, capturedCmd, 3)
		script := capturedCmd[2]
		assert.Contains(t, script, maliciousCommand, "The setup command should be present when explicitly allowed")
	})
}
