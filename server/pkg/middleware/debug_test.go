// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestDebugMiddleware(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var logOutput bytes.Buffer
	logging.Init(slog.LevelDebug, &logOutput)

	mw := DebugMiddleware()
	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "test-tool"},
			},
		}, nil
	})

	req := &mcp.ListToolsRequest{}
	_, err := handler(context.Background(), "tools/list", req)
	assert.NoError(t, err)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "method=tools/list")
	assert.Contains(t, logStr, "request")
	assert.Contains(t, logStr, "response")
	assert.Contains(t, logStr, "test-tool")
}

func TestDebugMiddleware_NoLoggingWhenDisabled(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var logOutput bytes.Buffer
	logging.Init(slog.LevelInfo, &logOutput)

	mw := DebugMiddleware()
	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "test-tool"},
			},
		}, nil
	})

	req := &mcp.ListToolsRequest{}
	_, err := handler(context.Background(), "tools/list", req)
	assert.NoError(t, err)

	assert.False(t, strings.Contains(logOutput.String(), "MCP Request"))
	assert.False(t, strings.Contains(logOutput.String(), "MCP Response"))
}

// mockUnmarshallableResult embeds a standard Result but adds a func to fail marshalling
type mockUnmarshallableResult struct {
	mcp.ListToolsResult
}

func (m *mockUnmarshallableResult) MarshalJSON() ([]byte, error) {
	return nil, &json.UnsupportedTypeError{Type: reflect.TypeOf(func() {})}
}

func TestDebugMiddleware_MarshalErrors(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var logOutput bytes.Buffer
	logging.Init(slog.LevelDebug, &logOutput)

	mw := DebugMiddleware()

	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mockUnmarshallableResult{
			ListToolsResult: mcp.ListToolsResult{},
		}, nil
	})

	req := &mcp.ListToolsRequest{}
	_, err := handler(context.Background(), "tools/list", req)
	assert.NoError(t, err)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "Failed to marshal response for debugging")
}

type unmarshallableRequest struct {
	mcp.ListToolsRequest
}

func (u *unmarshallableRequest) MarshalJSON() ([]byte, error) {
	return nil, &json.UnsupportedTypeError{Type: reflect.TypeOf(func() {})}
}

func TestDebugMiddleware_RequestMarshalError(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var logOutput bytes.Buffer
	logging.Init(slog.LevelDebug, &logOutput)

	mw := DebugMiddleware()
	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return nil, nil
	})

	// Pass unmarshallable request
	req := &unmarshallableRequest{
		ListToolsRequest: mcp.ListToolsRequest{},
	}
	_, err := handler(context.Background(), "test", req)
	// The handler should still succeed, just log error
	assert.NoError(t, err)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "Failed to marshal request for debugging")
}
