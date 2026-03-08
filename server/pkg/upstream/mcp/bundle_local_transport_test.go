// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestBundleLocalTransport_CaptureStderr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that prints to stderr and fails
	transport := &BundleLocalTransport{
		Command: "sh",
		Args:    []string{"-c", "echo 'something went wrong' >&2; exit 1"},
	}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Try to read. It should fail with EOF + Stderr info
	_, err = conn.Read(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "something went wrong")
}

func TestBundleLocalTransport_CaptureStderr_EarlyExit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport := &BundleLocalTransport{
		Command: "ls",
		Args:    []string{"/nonexistent_file_for_test"},
	}

	conn, err := transport.Connect(ctx)
	assert.NoError(t, err)
	defer conn.Close()

	// Try to read
	_, err = conn.Read(ctx)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "No such file") || strings.Contains(err.Error(), "ls:"), "Error should contain stderr output from ls")
}

func TestBundleLocalTransport_Success(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that acts as a dummy JSON-RPC server
	transport := &BundleLocalTransport{
		Command: "cat",
		Args:    []string{},
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

func TestBundleLocalTransport_ConnectFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command that does not exist
	transport := &BundleLocalTransport{
		Command: "/nonexistent_binary_for_test",
		Args:    []string{},
	}

	_, err := transport.Connect(ctx)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "no such file or directory"), "Error should indicate file not found")
}
