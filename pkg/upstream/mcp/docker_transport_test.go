/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

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
	conn := &dockerConn{rwc: rwc}

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

func TestSlogWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slogWriter := &slogWriter{log: log, level: slog.LevelInfo}

	testCases := []struct {
		name          string
		input         string
		expectedLines []map[string]any
	}{
		{
			name:  "single line",
			input: "hello world",
			expectedLines: []map[string]any{
				{"level": "INFO", "msg": "hello world"},
			},
		},
		{
			name:  "multiple lines",
			input: "line1\nline2",
			expectedLines: []map[string]any{
				{"level": "INFO", "msg": "line1"},
				{"level": "INFO", "msg": "line2"},
			},
		},
		{
			name:          "empty string",
			input:         "",
			expectedLines: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			_, err := slogWriter.Write([]byte(tc.input))
			assert.NoError(t, err)

			if len(tc.expectedLines) == 0 {
				assert.Equal(t, "", buf.String())
				return
			}

			output := strings.TrimSpace(buf.String())
			lines := strings.Split(output, "\n")
			require.Equal(t, len(tc.expectedLines), len(lines))

			for i, line := range lines {
				var actual map[string]any
				err := json.Unmarshal([]byte(line), &actual)
				require.NoError(t, err)

				expected := tc.expectedLines[i]
				assert.Equal(t, expected["level"], actual["level"])
				assert.Equal(t, expected["msg"], actual["msg"])
			}
		})
	}
}

func TestDockerTransport_Integration(t *testing.T) {
	if !util.IsDockerSocketAccessible() {
		t.Skip("Docker socket not accessible, skipping integration test.")
	}

	ctx := context.Background()
	transport := &DockerTransport{
		StdioConfig: configv1.McpStdioConnection_builder{
			Command:        proto.String("cat"),
			ContainerImage: proto.String("alpine:latest"),
		}.Build(),
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Get the container ID from the connection's ReadWriteCloser
	dockerRWC, ok := conn.(*dockerConn).rwc.(*dockerReadWriteCloser)
	require.True(t, ok, "connection should be a dockerReadWriteCloser")
	containerID := dockerRWC.containerID

	// Verify the container is running
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)
	defer cli.Close()

	_, err = cli.ContainerInspect(ctx, containerID)
	require.NoError(t, err, "container should be running")

	// Write to stdin and read from stdout
	go func() {
		err := conn.Write(ctx, &jsonrpc.Request{Method: "test"})
		assert.NoError(t, err)
	}()

	// Since we're using `cat`, the container will close the stream after reading the input.
	// We expect to get an EOF when reading.
	_, err = conn.Read(ctx)
	require.ErrorIs(t, err, io.EOF)

	// Close the connection and verify cleanup
	err = conn.Close()
	require.NoError(t, err)

	// Give the container time to be removed
	time.Sleep(2 * time.Second)

	// Verify the container is removed
	_, err = cli.ContainerInspect(ctx, containerID)
	require.True(t, client.IsErrNotFound(err), "container should be removed after close")
}
