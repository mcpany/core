// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

type threadSafeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *threadSafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *threadSafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func TestInspectorMiddleware(t *testing.T) {
	// Setup custom logger to capture output
	var buf threadSafeBuffer

	// Reset the logger singleton
	logging.ForTestsOnlyResetLogger()
	// Init with JSON format to buffer
	logging.Init(slog.LevelInfo, &buf, "json")

	// Define next handler that returns success
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{IsError: false}, nil
	}

	// Wrap with middleware
	middleware := InspectorMiddleware(next)

	// Call
	_, err := middleware(context.Background(), "tools/call", &mcp.CallToolRequest{})
	assert.NoError(t, err)

	// Wait a bit for async logging
	time.Sleep(200 * time.Millisecond)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "INSPECTOR")
	assert.Contains(t, output, "tools/call")
	assert.Contains(t, output, "payload")
}
