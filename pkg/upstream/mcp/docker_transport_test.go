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
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"testing"
	"time"

	"bufio"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
)

// mockDockerClient is a mock implementation of the Docker client API.
type mockDockerClient struct {
	client.APIClient
	imagePullFunc       func(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	containerCreateFunc func(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	containerAttachFunc func(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error)
	containerStartFunc  func(ctx context.Context, container string, options container.StartOptions) error
	containerStopFunc   func(ctx context.Context, container string, options container.StopOptions) error
	containerRemoveFunc func(ctx context.Context, container string, options container.RemoveOptions) error
	closeFunc           func() error
}

func (m *mockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	if m.imagePullFunc != nil {
		return m.imagePullFunc(ctx, ref, options)
	}
	return io.NopCloser(bytes.NewReader([]byte(""))), nil
}

func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if m.containerCreateFunc != nil {
		return m.containerCreateFunc(ctx, config, hostConfig, networkingConfig, platform, containerName)
	}
	return container.CreateResponse{ID: "test-container"}, nil
}

func (m *mockDockerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	if m.containerAttachFunc != nil {
		return m.containerAttachFunc(ctx, container, options)
	}
	return types.HijackedResponse{
		Reader: bufio.NewReader(bytes.NewReader(nil)),
		Conn:   &mockReadWriteCloser{},
	}, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	if m.containerStartFunc != nil {
		return m.containerStartFunc(ctx, container, options)
	}
	return nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.containerStopFunc != nil {
		return m.containerStopFunc(ctx, containerID, options)
	}
	return nil
}

func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.containerRemoveFunc != nil {
		return m.containerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

func (m *mockDockerClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

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

func (m *mockReadWriteCloser) Close() error                               { return nil }
func (m *mockReadWriteCloser) LocalAddr() net.Addr                        { return nil }
func (m *mockReadWriteCloser) RemoteAddr() net.Addr                       { return nil }
func (m *mockReadWriteCloser) SetDeadline(t time.Time) error              { return nil }
func (m *mockReadWriteCloser) SetReadDeadline(t time.Time) error          { return nil }
func (m *mockReadWriteCloser) SetWriteDeadline(t time.Time) error         { return nil }
func (m *mockReadWriteCloser) SetWriteBuffer(bytes int) error             { return nil }
func (m *mockReadWriteCloser) SetReadBuffer(bytes int) error              { return nil }
func (m *mockReadWriteCloser) SetKeepAlive(keepalive bool) error          { return nil }
func (m *mockReadWriteCloser) SetKeepAlivePeriod(d time.Duration) error   { return nil }
func (m *mockReadWriteCloser) SetNoDelay(noDelay bool) error              { return nil }
func (m *mockReadWriteCloser) SetLinger(sec int) error                    { return nil }
func (m *mockReadWriteCloser) SetMultipathTCP(a bool) error               { return nil }
func (m *mockReadWriteCloser) MultipathTCP() (bool, error)                { return false, nil }
func (m *mockReadWriteCloser) SyscallConn() (net.Conn, error)             { return nil, nil }
func (m *mockReadWriteCloser) File() (f *net.OpError, err error)          { return nil, nil }
func (m *mockReadWriteCloser) ReadFrom(r io.Reader) (n int64, err error)  { return 0, nil }
func (m *mockReadWriteCloser) WriteTo(w io.Writer) (n int64, err error)   { return 0, nil }
func (m *mockReadWriteCloser) CloseRead() error                           { return nil }
func (m *mockReadWriteCloser) CloseWrite() error                          { return nil }
func (m *mockReadWriteCloser) Connection() (net.Conn, error)              { return nil, nil }
func (m *mockReadWriteCloser) Handshake() error                           { return nil }
func (m *mockReadWriteCloser) HandshakeContext(ctx context.Context) error { return nil }
func (m *mockReadWriteCloser) ConnectionState() interface{}               { return nil }

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

func TestDockerTransport_Connect_Integration(t *testing.T) {
	t.Skip("Skipping integration test")
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("printf")
	stdioConfig.SetArgs([]string{`'{"jsonrpc": "2.0", "id": "1", "result": "hello"}'`})
	transport := &DockerTransport{StdioConfig: stdioConfig}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	msg, err := conn.Read(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, msg)

	resp, ok := msg.(*jsonrpc.Response)
	assert.True(t, ok)
	assert.Equal(t, "1", resp.ID.Raw())
	assert.Equal(t, json.RawMessage(`"hello"`), resp.Result)

	err = conn.Close()
	assert.NoError(t, err)
}

func TestDockerTransport_Connect_ImageNotFound(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("this-image-does-not-exist-ever:latest")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
}

func TestDockerTransport_Connect_NoImage(t *testing.T) {
	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container_image must be specified")
}

func TestDockerReadWriteCloser_Close(t *testing.T) {
	mockClient := &mockDockerClient{}
	var stopCalled, removeCalled bool
	mockClient.containerStopFunc = func(ctx context.Context, container string, options container.StopOptions) error {
		stopCalled = true
		return nil
	}
	mockClient.containerRemoveFunc = func(ctx context.Context, container string, options container.RemoveOptions) error {
		removeCalled = true
		return nil
	}

	rwc := &dockerReadWriteCloser{
		WriteCloser: &mockReadWriteCloser{},
		containerID: "test-container",
		cli:         mockClient,
	}

	err := rwc.Close()
	assert.NoError(t, err)
	assert.True(t, stopCalled, "ContainerStop should have been called")
	assert.True(t, removeCalled, "ContainerRemove should have been called")
}

func TestDockerTransport_Connect_ContainerCreateFail(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
}

func TestDockerTransport_Connect_ContainerAttachFail(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}
	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
}

func TestDockerTransport_Connect_ContainerStartFail(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}
	ctx := context.Background()
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
}
