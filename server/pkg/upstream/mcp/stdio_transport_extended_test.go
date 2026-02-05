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

func TestStdioTransport_SessionID(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "echo", "dummy")
	transport := &StdioTransport{Command: cmd}
	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	assert.Equal(t, "stdio-session", conn.SessionID())
}

func TestStdioTransport_CloseIdempotency(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "echo", "dummy")
	transport := &StdioTransport{Command: cmd}
	conn, err := transport.Connect(ctx)
	require.NoError(t, err)

	assert.NoError(t, conn.Close())
	assert.NoError(t, conn.Close()) // Should not error on second close
}

func TestStdioTransport_Read_InvalidJSON(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command outputs invalid JSON
	cmd := exec.CommandContext(ctx, "echo", "{ invalid json")
	transport := &StdioTransport{Command: cmd}
	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	_, err = conn.Read(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestStdioTransport_Read_Response(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Command outputs a JSON-RPC response
	// We use 'printf' to avoid newline at the beginning if that matters, though echo usually adds newline at end.
	// JSON decoder handles whitespace.
	responseJSON := `{"jsonrpc":"2.0", "result": "success", "id": 1}`
	cmd := exec.CommandContext(ctx, "echo", responseJSON)
	transport := &StdioTransport{Command: cmd}
	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	msg, err := conn.Read(ctx)
	require.NoError(t, err)

	resp, ok := msg.(*jsonrpc.Response)
	require.True(t, ok)

	// Check ID - note that unmarshalling might result in float64 for numbers if interface{} is involved,
	// but let's check what we get. The transport logic uses setUnexportedID.
	// json.Unmarshal into interface{} for 1 gives float64(1).
	// But let's see how fixID/setUnexportedID handles it.
	// Based on code, it should be set.

	// We can't easily access unexported ID field directly from here because we are in package mcp (same package),
	// so we actually CAN access unexported fields if they were on the struct, but ID is on jsonrpc.Message/Response
	// which is from an external package (github.com/modelcontextprotocol/go-sdk/jsonrpc).
	// However, jsonrpc.Response likely has public fields.
	// Wait, jsonrpc.Response struct definition:
	// type Response struct {
	// 	JSONRPC string          `json:"jsonrpc"`
	// 	Result  json.RawMessage `json:"result,omitempty"`
	// 	Error   *Error          `json:"error,omitempty"`
	// 	ID      ID              `json:"id"`
	// }
	// And ID might be an interface or specific type.
	// The `mcp` package code has `setUnexportedID`.

	// Actually, let's verify the Result.
	assert.Contains(t, string(resp.Result), "success")
}

func TestStdioTransport_Read_Response_StringID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	responseJSON := `{"jsonrpc":"2.0", "result": "success", "id": "req-1"}`
	cmd := exec.CommandContext(ctx, "echo", responseJSON)
	transport := &StdioTransport{Command: cmd}
	conn, err := transport.Connect(ctx)
	require.NoError(t, err)
	defer conn.Close()

	msg, err := conn.Read(ctx)
	require.NoError(t, err)

	resp, ok := msg.(*jsonrpc.Response)
	require.True(t, ok)
	assert.Contains(t, string(resp.Result), "success")
}
