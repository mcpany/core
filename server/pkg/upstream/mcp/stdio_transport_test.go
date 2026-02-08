package mcp

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
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
