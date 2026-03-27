// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdioTransport_CaptureStderr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that prints to stderr and fails
	cmd := exec.CommandContext(ctx, "sh", "-c", "echo 'something went wrong' >&2; exit 1")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Try to read. It should fail with EOF + Stderr info
	_, err = conn.Read(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "something went wrong")
}

func TestStdioTransport_CaptureStderr_EarlyExit(t *testing.T) {
	// Tests the case where the process exits immediately but maybe without stderr or with minimal
	// This mirrors the "ls /nonexistent" case
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ls", "/nonexistent_file_for_test")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Try to read
	_, err = conn.Read(ctx)
	assert.Error(t, err)
	// Output depends on OS/locale but usually contains "No such file"
	assert.True(t, strings.Contains(err.Error(), "No such file") || strings.Contains(err.Error(), "ls:"), "Error should contain stderr output from ls")
}

func TestStdioTransport_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that acts as a dummy JSON-RPC server
	// We just echo a message
	cmd := exec.CommandContext(ctx, "cat")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Write a message
	req := &jsonrpc.Request{
		Method: "ping",
	}
	setUnexportedID(&req.ID, 1)
	err = conn.Write(ctx, req)
	assert.NoError(t, err)

	// Read it back
	msg, err := conn.Read(ctx)
	assert.NoError(t, err)

	reqRead, ok := msg.(*jsonrpc.Request)
	assert.True(t, ok)
	assert.Equal(t, "ping", reqRead.Method)
}

func TestStdioTransport_ConnectAndReadWrite(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use `cat` as a mock command that just echoes what it receives
	cmd := exec.CommandContext(ctx, "cat")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	assert.Equal(t, "stdio-session", conn.SessionID())

	// Write a message
	msg := &jsonrpc.Request{
		Method: "test/method",
		Params: []byte(`{"foo":"bar"}`),
	}

	err = conn.Write(ctx, msg)
	require.NoError(t, err)

	// Read the message back
	readCh := make(chan jsonrpc.Message)
	errCh := make(chan error)

	go func() {
		readMsg, err := conn.Read(ctx)
		if err != nil {
			errCh <- err
			return
		}
		readCh <- readMsg
	}()

	select {
	case readMsg := <-readCh:
		req, ok := readMsg.(*jsonrpc.Request)
		assert.True(t, ok)
		assert.Equal(t, msg.Method, req.Method)
		assert.JSONEq(t, string(msg.Params), string(req.Params))
	case err := <-errCh:
		t.Fatalf("Failed to read message: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout reading message")
	}
}

func TestStdioTransport_ConnectError(t *testing.T) {
	ctx := context.Background()

	// Use a non-existent command
	cmd := exec.CommandContext(ctx, "non-existent-command-12345")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestStdioTransport_Close(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "sleep", "10")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)

	// Verify command is running
	require.NotNil(t, cmd.Process)

	err = conn.Close()
	assert.NoError(t, err)

	// Give it a moment to terminate
	time.Sleep(100 * time.Millisecond)

	// Process should be done
	err = cmd.Wait()
	assert.Error(t, err) // It was killed, so wait returns an error
}

func TestStdioTransport_ReadBadJSON(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "echo", "not json")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	// Read should fail
	_, err = conn.Read(ctx)
	assert.Error(t, err)
}
