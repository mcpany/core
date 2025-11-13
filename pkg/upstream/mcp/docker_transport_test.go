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
	"io"
	"log/slog"
	"testing"

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
	"github.com/stretchr/testify/mock"
)

type MockDockerClient struct {
	client.APIClient
	mock.Mock
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
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

func (m *mockReadWriteCloser) Close() error {
	return nil
}

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

func TestDockerTransport_Connect_ContainerCreateError(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockDockerClient)
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{
		StdioConfig: stdioConfig,
		NewClient: func() (client.APIClient, error) {
			return mockClient, nil
		},
	}

	mockClient.On("ImagePull", ctx, "test-image", mock.Anything).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil)
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.CreateResponse{}, assert.AnError)

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestDockerTransport_Connect_ContainerAttachError(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockDockerClient)
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{
		StdioConfig: stdioConfig,
		NewClient: func() (client.APIClient, error) {
			return mockClient, nil
		},
	}

	mockClient.On("ImagePull", ctx, "test-image", mock.Anything).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil)
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.CreateResponse{ID: "test-id"}, nil)
	mockClient.On("ContainerAttach", ctx, "test-id", mock.Anything).Return(types.HijackedResponse{}, assert.AnError)
	mockClient.On("ContainerRemove", ctx, "test-id", mock.Anything).Return(nil)

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestDockerTransport_Connect_ContainerStartError(t *testing.T) {
	ctx := context.Background()
	mockClient := new(MockDockerClient)
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("test-image")
	stdioConfig.SetCommand("echo")
	transport := &DockerTransport{
		StdioConfig: stdioConfig,
		NewClient: func() (client.APIClient, error) {
			return mockClient, nil
		},
	}

	mockClient.On("ImagePull", ctx, "test-image", mock.Anything).Return(io.NopCloser(bytes.NewReader([]byte(""))), nil)
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.CreateResponse{ID: "test-id"}, nil)
	mockClient.On("ContainerAttach", ctx, "test-id", mock.Anything).Return(types.HijackedResponse{}, nil)
	mockClient.On("ContainerStart", ctx, "test-id", mock.Anything).Return(assert.AnError)
	mockClient.On("ContainerRemove", ctx, "test-id", mock.Anything).Return(nil)

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}
