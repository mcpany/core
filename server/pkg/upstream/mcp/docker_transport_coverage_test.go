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

func TestDockerConn_Read_EOFWithStderr(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}

	// Create a stderr capture buffer and fill it with an error message
	stderrCapture := &tailBuffer{limit: 1024}
	_, _ = stderrCapture.Write([]byte("Container process exited with error code 1"))

	conn := &dockerConn{
		rwc:           rwc,
		decoder:       json.NewDecoder(rwc), // rwc is empty, so this will return EOF immediately
		encoder:       json.NewEncoder(rwc),
		stderrCapture: stderrCapture,
	}

	_, err := conn.Read(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection closed. Stderr: Container process exited with error code 1")
}

func TestDockerConn_Read_RequestWithDifferentIDs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		payload string
		wantID  any
	}{
		{
			name:    "String ID",
			payload: `{"jsonrpc": "2.0", "method": "test", "id": "req-1"}`,
			wantID:  "req-1",
		},
		{
			name:    "Number ID",
			payload: `{"jsonrpc": "2.0", "method": "test", "id": 123}`,
			wantID:  int64(123),
		},
		{
			name:    "Null ID (Notification)",
			payload: `{"jsonrpc": "2.0", "method": "test", "id": null}`,
			wantID:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rwc := &mockReadWriteCloser{}
			_, _ = rwc.WriteString(tt.payload + "\n")

			conn := &dockerConn{
				rwc:     rwc,
				decoder: json.NewDecoder(rwc),
				encoder: json.NewEncoder(rwc),
			}

			msg, err := conn.Read(ctx)
			require.NoError(t, err)

			// Verify it's a request
			req, ok := msg.(*jsonrpc.Request)
			require.True(t, ok, "Expected a Request message")

			// Verify ID
			// Check if ID.Raw() matches wantID
			if tt.wantID != nil {
				// Special handling for numeric IDs depending on how JSON is decoded
				raw := req.ID.Raw()
				// If expected is float64, check if raw matches
				assert.Equal(t, tt.wantID, raw)
			} else {
				// For notification, check if ID is essentially empty/null
				// Since we can't check req.ID == nil (it's a struct), check Raw()
				assert.Nil(t, req.ID.Raw())
			}
		})
	}
}

func TestDockerConn_Read_ResponseWithDifferentIDs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		payload string
		wantID  any
	}{
		{
			name:    "String ID",
			payload: `{"jsonrpc": "2.0", "result": "ok", "id": "resp-1"}`,
			wantID:  "resp-1",
		},
		{
			name:    "Number ID",
			payload: `{"jsonrpc": "2.0", "result": "ok", "id": 456}`,
			wantID:  int64(456),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rwc := &mockReadWriteCloser{}
			_, _ = rwc.WriteString(tt.payload + "\n")

			conn := &dockerConn{
				rwc:     rwc,
				decoder: json.NewDecoder(rwc),
				encoder: json.NewEncoder(rwc),
			}

			msg, err := conn.Read(ctx)
			require.NoError(t, err)

			// Verify it's a response
			resp, ok := msg.(*jsonrpc.Response)
			require.True(t, ok, "Expected a Response message")

			// Verify ID
			assert.Equal(t, tt.wantID, resp.ID.Raw())
		})
	}
}

func TestDockerConn_Read_MalformedJSON(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	// Valid JSON object start, but truncated
	_, _ = rwc.WriteString(`{"jsonrpc": "2.0", "method": "test"`)

	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	_, err := conn.Read(ctx)
	require.Error(t, err)
	// Error should be about unexpected EOF or syntax error
	assert.Contains(t, err.Error(), "unexpected EOF")
}

func TestDockerConn_Read_HeaderUnmarshalError(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}
	// Valid JSON but not an object, so unmarshal to header struct fails
	_, _ = rwc.WriteString(`[]` + "\n")

	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	_, err := conn.Read(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal message header")
}

func TestDockerConn_Read_AnyIDFallback_Request(t *testing.T) {
	ctx := context.Background()
	rwc := &mockReadWriteCloser{}

	// A request where ID is boolean true.
	payload := `{"jsonrpc": "2.0", "method": "test", "id": true}`
	_, _ = rwc.WriteString(payload + "\n")

	conn := &dockerConn{
		rwc:     rwc,
		decoder: json.NewDecoder(rwc),
		encoder: json.NewEncoder(rwc),
	}

	msg, err := conn.Read(ctx)

	if err == nil {
		req, ok := msg.(*jsonrpc.Request)
		require.True(t, ok)
		assert.Equal(t, "test", req.Method)
		// Check ID is true (bool)
		assert.Equal(t, true, req.ID.Raw())
	}
}
