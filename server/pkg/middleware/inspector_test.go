package middleware

import (
	"context"
	"testing"
	"time"
	"bytes"
	"log/slog"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestInspectorMiddleware(t *testing.T) {
	// Setup custom logger to capture output
	var buf bytes.Buffer

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
