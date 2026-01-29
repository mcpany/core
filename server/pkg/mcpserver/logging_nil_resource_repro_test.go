package mcpserver_test

import (
	"testing"
	"log/slog"
	"bytes"
	"strings"

	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestLazyLogResult_PanicRepro(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()
	var logBuffer bytes.Buffer
	logging.Init(slog.LevelInfo, &logBuffer, "text")
	defer func() {
		logging.ForTestsOnlyResetLogger()
		logging.GetLogger()
	}()

	// Create a CallToolResult with EmbeddedResource having nil Resource
	badResult := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.EmbeddedResource{
				Resource: nil, // This should cause a panic in summarizeCallToolResult
			},
		},
	}

	logger := logging.GetLogger()
	logger.Info("Testing logging", "result", mcpserver.LazyLogResult{Value: badResult})

	// Check if log contains "PANIC" (slog default behavior when LogValuer panics)
	logOutput := logBuffer.String()
	if strings.Contains(logOutput, "PANIC") || strings.Contains(logOutput, "panic") {
		t.Fatalf("Log output contained panic: %s", logOutput)
	}
}
