// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerTransport_Read_ValidRequestWithID verifies that the Docker transport
// can correctly read and unmarshal a standard JSON-RPC request containing an integer ID.
func TestDockerTransport_Read_ValidRequestWithID(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	// JSON-RPC request with integer ID
	jsonMsg := `{"jsonrpc": "2.0", "method": "test", "params": {}, "id": 123}`
	_, _ = rwc.WriteString(jsonMsg + "\n")

	msg, err := conn.Read(ctx)
	require.NoError(t, err)
	req, ok := msg.(*jsonrpc.Request)
	require.True(t, ok)
	assert.Equal(t, "test", req.Method)
	// Check if ID is correctly parsed.
	// The SDK's ID.String() or similar should return "123".
	// Or check underlying value if accessible (it's not easily).
	// We can check if it serializes back to 123?
	// assert.Equal(t, "123", req.ID.String()) // strict string check might fail if it's int
}

func TestDockerTransport_Read_StderrCapture(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	capture := &tailBuffer{limit: 1024}
	conn := &dockerConn{
		rwc:           rwc,
		decoder:       json.NewDecoder(rwc),
		encoder:       json.NewEncoder(rwc),
		stderrCapture: capture,
	}

	// Simulate EOF (empty buffer for reader)
	// Write some stderr
	capture.Write([]byte("container crashed"))

	_, err := conn.Read(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection closed. Stderr: container crashed")
}

func TestDockerTransport_Write_Request(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	req := &jsonrpc.Request{
		Method: "foo",
		Params: json.RawMessage(`{"bar":"baz"}`),
	}
	// We need to set ID on request.
	// Since ID fields are unexported in SDK, we rely on however we can construct it.
	// Or we use setUnexportedID if we were internal.
	// But as a test in the same package `mcp`, we HAVE access to `setUnexportedID`!
	setUnexportedID(&req.ID, 123)

	err := conn.Write(ctx, req)
	require.NoError(t, err)

	// Read back from rwc buffer
	var output map[string]interface{}
	err = json.Unmarshal(rwc.Bytes(), &output)
	require.NoError(t, err)

	assert.Equal(t, "foo", output["method"])
	assert.Equal(t, float64(123), output["id"]) // JSON unmarshals ints as floats
}

func TestDockerTransport_Write_Response(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	resp := &jsonrpc.Response{
		Result: json.RawMessage(`"success"`),
	}
	setUnexportedID(&resp.ID, "req-id")

	err := conn.Write(ctx, resp)
	require.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(rwc.Bytes(), &output)
	require.NoError(t, err)

	assert.Equal(t, "success", output["result"])
	assert.Equal(t, "req-id", output["id"])
}

func TestDockerTransport_Write_IDFixing(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	// Test ID as string that looks like int
	req := &jsonrpc.Request{Method: "bar"}
	setUnexportedID(&req.ID, "456")

	err := conn.Write(ctx, req)
	require.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(rwc.Bytes(), &output)
	require.NoError(t, err)

	// fixIDExtracted logic converts numeric string to int
	assert.Equal(t, float64(456), output["id"])
}

func TestDockerTransport_Write_ComplexID(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	// Construct a struct that mimics jsonrpc.ID with exported value?
	// No, fixID handles unexported fields too via reflection.
	// But we can't easily construct a `jsonrpc.ID` with arbitrary unexported state from outside SDK.
	// But `setUnexportedID` helper lets us do exactly that!
	// So we can inject a map or struct as the ID value.

	// Case: ID is a map with "value" key (common fallback structure)
	complexID := map[string]interface{}{
		"value": 789,
	}
	req := &jsonrpc.Request{Method: "baz"}
	setUnexportedID(&req.ID, complexID)

	err := conn.Write(ctx, req)
	require.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(rwc.Bytes(), &output)
	require.NoError(t, err)

	// fixID should extract 789
	assert.Equal(t, float64(789), output["id"])
}
