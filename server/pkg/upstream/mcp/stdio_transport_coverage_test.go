// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStdioTransport_Read_Request_StringID ...
// Summary: TestStdioTransport_Read_Request_StringID
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Request with string ID
	cmd := exec.CommandContext(ctx, "printf", `{"jsonrpc":"2.0","method":"ping","id":"string_id"}`)

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	msg, err := conn.Read(ctx)
	require.NoError(t, err)

	req, ok := msg.(*jsonrpc.Request)
	require.True(t, ok)
	assert.Equal(t, "ping", req.Method)
}

// TestStdioTransport_Read_HeaderUnmarshalError ...
// Summary: TestStdioTransport_Read_HeaderUnmarshalError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Valid JSON but not an object (array), which fails header unmarshal
	cmd := exec.CommandContext(ctx, "printf", `["not", "an", "object"]`)

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Read(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message header")
}

// TestStdioTransport_Write_StringID ...
// Summary: TestStdioTransport_Write_StringID
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Echo server
	cmd := exec.CommandContext(ctx, "cat")

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	req := &jsonrpc.Request{
		Method: "test",
	}

	// Use helper from bundle_transport.go (same package)
	err = setUnexportedID(&req.ID, "my-string-id")
	require.NoError(t, err)

	err = conn.Write(ctx, req)
	require.NoError(t, err)

	// Verify what was written by reading it back
	msg, err := conn.Read(ctx)
	require.NoError(t, err)

	reqRead, ok := msg.(*jsonrpc.Request)
	require.True(t, ok)
	assert.Equal(t, "test", reqRead.Method)
}

// TestStdioTransport_Read_RequestUnmarshalError ...
// Summary: TestStdioTransport_Read_RequestUnmarshalError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Request with object ID (complex ID) to trigger fallback if standard unmarshal fails
	cmd := exec.CommandContext(ctx, "printf", `{"jsonrpc":"2.0","method":"ping","id":{"complex":"id"}}`)

	transport := &StdioTransport{
		Command: cmd,
	}

	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	msg, err := conn.Read(ctx)
	require.NoError(t, err)

	req, ok := msg.(*jsonrpc.Request)
	require.True(t, ok)
	assert.Equal(t, "ping", req.Method)
}
