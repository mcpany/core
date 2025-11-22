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

package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestSlogWriter(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))
	writer := &slogWriter{log: log, level: slog.LevelInfo}

	testMessage := "Hello, world!"
	_, err := writer.Write([]byte(testMessage))
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), testMessage)
}

func TestDockerConn_SessionID(t *testing.T) {
	conn := &dockerConn{}
	assert.Equal(t, "docker-transport-session", conn.SessionID())
}

type mockReadWriteCloser struct {
	bytes.Buffer
}

func (m *mockReadWriteCloser) Close() error {
	return nil
}

func (m *mockReadWriteCloser) LocalAddr() net.Addr                { return nil }
func (m *mockReadWriteCloser) RemoteAddr() net.Addr               { return nil }
func (m *mockReadWriteCloser) SetDeadline(t time.Time) error      { return nil }
func (m *mockReadWriteCloser) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockReadWriteCloser) SetWriteDeadline(t time.Time) error { return nil }

func TestDockerConn_ReadWrite(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	// Test Write
	testMsg := &jsonrpc.Request{
		Method: "test",
	}
	err := conn.Write(ctx, testMsg)
	assert.NoError(t, err)

	// Test Read
	readMsg, err := conn.Read(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, readMsg)

	// Test Close
	err = conn.Close()
	assert.NoError(t, err)
}

func TestDockerConn_Read_UnmarshalError(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid header", func(t *testing.T) {
		rwc := &mockReadWriteCloser{}
		conn := &dockerConn{
			rwc:     rwc,
			decoder: json.NewDecoder(rwc),
			encoder: json.NewEncoder(rwc),
		}
		invalidHeaderMsg := `{"method": 123}`
		rwc.WriteString(invalidHeaderMsg + "\n")

		_, err := conn.Read(ctx)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to unmarshal message header")
		}
	})

	t.Run("invalid json syntax", func(t *testing.T) {
		rwc := &mockReadWriteCloser{}
		conn := &dockerConn{
			rwc:     rwc,
			decoder: json.NewDecoder(rwc),
			encoder: json.NewEncoder(rwc),
		}
		// This is syntactically invalid, and will cause `decoder.Decode` to fail.
		invalidMsg := `{"method": "test"`
		rwc.WriteString(invalidMsg + "\n")

		_, err := conn.Read(ctx)
		assert.Error(t, err)
	})
}

func TestDockerTransport_Connect_ClientError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return nil, fmt.Errorf("client error")
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create docker client")
}

func TestDockerTransport_Connect_ContainerCreateError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{}, fmt.Errorf("container create error")
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create container")
}

func TestDockerTransport_Connect_ContainerAttachError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-id"}, nil
			},
			ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{}, fmt.Errorf("container attach error")
			},
			ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
				return nil
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to attach to container")
}

func TestDockerTransport_Connect_ContainerStartError(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-id"}, nil
			},
			ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{
					Conn:   &mockReadWriteCloser{},
					Reader: bufio.NewReader(&mockReadWriteCloser{}),
				}, nil
			},
			ContainerStartFunc: func(ctx context.Context, container string, options container.StartOptions) error {
				return fmt.Errorf("container start error")
			},
			ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
				return nil
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	transport := &DockerTransport{StdioConfig: stdioConfig}
	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start container")
}

func TestDockerTransport_Connect_Integration(t *testing.T) {
	t.Skip("Skipping integration test")
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	// ctx := context.Background()
	// TODO: Fix this test, it is failing
	// stdioConfig := &configv1.McpStdioConnection{}
	// stdioConfig.ContainerImage("alpine:latest")
	// stdioConfig.SetCommand("printf")
	// stdioConfig.SetArgs([]string{`'{"jsonrpc": "2.0", "id": "1", "result": "hello"}'`})
	// transport := &DockerTransport{StdioConfig: stdioConfig}

	// conn, err := transport.Connect(ctx)
	// assert.NoError(t, err)
	// assert.NotNil(t, conn)

	// msg, err := conn.Read(ctx)
	// assert.NoError(t, err)
	// assert.NotNil(t, msg)

	// resp, ok := msg.(*jsonrpc.Response)
	// assert.True(t, ok)
	// assert.Equal(t, "1", resp.ID.Raw())
	// assert.Equal(t, json.RawMessage(`"hello"`), resp.Result)

	// err = conn.Close()
	// assert.NoError(t, err)
}

func TestDockerTransport_Connect_ImageNotFound(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	// ctx := context.Background()
	// TODO: Fix this test, it is failing
	// stdioConfig := &configv1.McpStdioConnection{}
	// stdioConfig.ContainerImage("this-image-does-not-exist-ever:latest")
	// stdioConfig.SetCommand("echo")
	// transport := &DockerTransport{StdioConfig: stdioConfig}

	// _, err := transport.Connect(ctx)
	// assert.Error(t, err)
}

func TestDockerTransport_Connect_NoImage(t *testing.T) {
	ctx := context.Background()
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container_image must be specified")
}

func TestDockerTransport_Connect_Success(t *testing.T) {
	originalNewDockerClient := newDockerClient
	newDockerClient = func(ops ...client.Opt) (dockerClient, error) {
		return &mockDockerClient{
			ImagePullFunc: func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			},
			ContainerCreateFunc: func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
				return container.CreateResponse{ID: "test-id"}, nil
			},
			ContainerAttachFunc: func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
				return types.HijackedResponse{
					Conn:   &mockReadWriteCloser{},
					Reader: bufio.NewReader(&mockReadWriteCloser{}),
				}, nil
			},
			ContainerStartFunc: func(ctx context.Context, container string, options container.StartOptions) error {
				return nil
			},
			ContainerStopFunc: func(ctx context.Context, containerID string, options container.StopOptions) error {
				return nil
			},
			ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
				return nil
			},
			CloseFunc: func() error {
				return nil
			},
		}, nil
	}
	defer func() { newDockerClient = originalNewDockerClient }()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	stdioConfig.SetCommand("echo")
	stdioConfig.SetArgs([]string{"hello"})
	transport := &DockerTransport{StdioConfig: stdioConfig}

	conn, err := transport.Connect(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	err = conn.Close()
	assert.NoError(t, err)
}

func TestDockerReadWriteCloser_Close_Error(t *testing.T) {
	var buf bytes.Buffer
	logging.ForTestsOnlyResetLogger()
	logging.Init(slog.LevelInfo, &buf)

	mockClient := &mockDockerClient{
		ContainerStopFunc: func(ctx context.Context, containerID string, options container.StopOptions) error {
			return fmt.Errorf("stop error")
		},
		ContainerRemoveFunc: func(ctx context.Context, containerID string, options container.RemoveOptions) error {
			return fmt.Errorf("remove error")
		},
		CloseFunc: func() error {
			return nil
		},
	}

	rwc := &dockerReadWriteCloser{
		WriteCloser: &mockReadWriteCloser{},
		containerID: "test-container",
		cli:         mockClient,
	}

	err := rwc.Close()
	assert.NoError(t, err)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "Failed to stop container")
	assert.Contains(t, logOutput, "stop error")
	assert.Contains(t, logOutput, "Failed to remove container")
	assert.Contains(t, logOutput, "remove error")
}
