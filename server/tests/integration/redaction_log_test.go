package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestRedactionInLogs(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()

	logBuffer := &threadSafeBuffer{}
	logging.Init(slog.LevelInfo, logBuffer)

	// Reset logger cleanup
	defer logging.ForTestsOnlyResetLogger()

	apiKey := "test-api-key"
	// Use the refactored helper with API key
	serverInfo := StartInProcessMCPANYServer(t, "redaction-test", apiKey)
	defer serverInfo.CleanupFunc()

	// Create a dummy upstream service
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintln(w, `{"content": [{"type":"text", "text":"ok"}]}`)
	}))
	defer upstream.Close()

	serviceName := "test-service"
	operationID := "test_op"

	// Register HTTP service
	RegisterHTTPService(t, serverInfo.RegistrationClient, serviceName, upstream.URL, operationID, "/", "GET", nil)

	require.NoError(t, serverInfo.Initialize(context.Background()))

	// Wait for tool to be available
	require.Eventually(t, func() bool {
		listToolsResult, err := serverInfo.ListTools(context.Background())
		if err != nil {
			t.Logf("ListTools error: %v", err)
			return false
		}
		var names []string
		for _, tool := range listToolsResult.Tools {
			names = append(names, tool.Name)
			if tool.Name == serviceName+"."+operationID {
				return true
			}
		}
		t.Logf("Waiting for tool. Found tools: %v", names)
		return false
	}, 10*time.Second, 500*time.Millisecond, "tool was not registered")

	// Call the tool with sensitive data
	callToolParams := &mcp.CallToolParams{
		Name:      serviceName + "." + operationID,
		Arguments: json.RawMessage(`{"credentials": "SECRET_VALUE_LEAKED", "other": "value"}`),
	}

	_, err := serverInfo.CallTool(context.Background(), callToolParams)
	require.NoError(t, err)

	// Check logs
	logs := logBuffer.String()

	// We expect "Calling tool..." log message
	if !strings.Contains(logs, "Calling tool...") {
		t.Logf("Full logs:\n%s", logs)
		t.Fatal("Expected log message 'Calling tool...' not found")
	}

	// We expect "credentials" key to be present (escaped or not)
	if !strings.Contains(logs, "credentials") {
		t.Logf("Full logs:\n%s", logs)
		t.Fatal("Expected 'credentials' key in logs")
	}

	// We expect "SECRET_VALUE_LEAKED" to be REDACTED
	if strings.Contains(logs, "SECRET_VALUE_LEAKED") {
		t.Fatalf("Security Leak! 'SECRET_VALUE_LEAKED' found in logs: %s", logs)
	}

	// We expect [REDACTED] to be present
	if !strings.Contains(logs, "[REDACTED]") {
		t.Fatalf("Expected '[REDACTED]' in logs, but not found. Logs snippet: %s", logs)
	}
}
