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
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestDockerTransport_SuccessfulLifecycle covers the successful creation, connection, and cleanup of a Docker container.
func TestDockerTransport_SuccessfulLifecycle(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("sh")
	stdioConfig.SetArgs([]string{"-c", `echo '{"jsonrpc":"2.0","result":"hello","id":1}' && sleep 5`})

	transport := &DockerTransport{StdioConfig: stdioConfig}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Verify that the container is running
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)
	defer cli.Close()

	dc, ok := conn.(*dockerConn)
	require.True(t, ok)
	drwc, ok := dc.rwc.(*dockerReadWriteCloser)
	require.True(t, ok)

	_, err = cli.ContainerInspect(ctx, drwc.containerID)
	require.NoError(t, err)

	// Read from the connection
	msg, err := conn.Read(ctx)
	require.NoError(t, err)
	require.NotNil(t, msg)

	resp, ok := msg.(*jsonrpc.Response)
	require.True(t, ok)
	assert.Equal(t, json.RawMessage(`"hello"`), resp.Result)

	// Close the connection and verify cleanup
	err = conn.Close()
	assert.NoError(t, err)

	// Check that the container is removed
	_, err = cli.ContainerInspect(ctx, drwc.containerID)
	assert.True(t, client.IsErrNotFound(err), "container should be removed after close")
}

// TestDockerTransport_CleanupOnError ensures that the container is cleaned up if an error occurs during connection.
func TestDockerTransport_CleanupOnError(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// This command will fail, causing cli.ContainerAttach to error out
	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("alpine:latest")
	stdioConfig.SetCommand("nonexistentcommand")

	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	require.Error(t, err)

	// Verify that no containers are left running
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	require.NoError(t, err)

	for _, c := range containers {
		assert.NotContains(t, c.Image, "alpine:latest", "no container with the test image should be running")
	}
}

// TestSlogWriter_StderrCapture verifies that stderr from the container is correctly captured and logged.
func TestSlogWriter_StderrCapture(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
	writer := &slogWriter{log: log, level: slog.LevelError}

	testMessage := "this is an error message"
	_, err := writer.Write([]byte(testMessage))
	require.NoError(t, err)

	assert.Contains(t, buf.String(), testMessage)
	assert.Contains(t, buf.String(), `"level":"ERROR"`)
}

// TestDockerTransport_ImagePullError tests the case where the Docker image pull fails.
func TestDockerTransport_ImagePullError(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	stdioConfig := &configv1.McpStdioConnection{}
	stdioConfig.SetContainerImage("nonexistentimage:latest")
	stdioConfig.SetCommand("echo")

	transport := &DockerTransport{StdioConfig: stdioConfig}

	_, err := transport.Connect(ctx)
	// We expect an error, but not a fatal one, as the code will try to use a local image.
	// The container creation will fail, and that's the error we're looking for.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create container")
}
